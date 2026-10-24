terraform {
  required_version = ">= 1.9.0"

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.45.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.75.0"
    }
  }

  backend "s3" {
    bucket         = "gvp-terraform-state-prod"
    key            = "global-video-platform/prod/terraform.tfstate"
    region         = "ap-northeast-2"
    encrypt        = true
    dynamodb_table = "gvp-terraform-locks"
  }
}

# ==============================================================================
# VARIABLES
# ==============================================================================
variable "environment" {
  type    = string
  default = "production"
}

variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare Account ID for R2, Stream, and CDN"
}

variable "cloudflare_zone_id" {
  type        = string
  description = "Cloudflare Zone ID for brand.com"
}

variable "aws_region" {
  type    = string
  default = "ap-northeast-2"
}

# ==============================================================================
# 1. CLOUDFLARE R2 BUCKETS (RAW UPLOADS, PUBLIC HLS, ARCHIVE/AUDIT)
# ==============================================================================
resource "cloudflare_r2_bucket" "raw_uploads" {
  account_id = var.cloudflare_account_id
  name       = "gvp-${var.environment}-raw-uploads"
  location   = "APAC"
}

resource "cloudflare_r2_bucket" "public_hls" {
  account_id = var.cloudflare_account_id
  name       = "gvp-${var.environment}-public-hls"
  location   = "APAC"
}

resource "cloudflare_r2_bucket" "audit_archive" {
  account_id = var.cloudflare_account_id
  name       = "gvp-${var.environment}-audit-archive"
  location   = "APAC"
}

# ==============================================================================
# 2. CLOUDFLARE CDN SIGNED URL POLICY & WAF RULES (media.brand.com)
# ==============================================================================
resource "cloudflare_ruleset" "media_cdn_signed_url_policy" {
  zone_id     = var.cloudflare_zone_id
  name        = "GVP Media CDN Signed Token Verification"
  description = "Enforce HMAC-SHA256 signed URL tokens and HLS segment expiration on media.brand.com"
  kind        = "zone"
  phase       = "http_request_firewall_custom"

  rules {
    action      = "block"
    expression  = "(http.host eq \"media.brand.com\" and not is_timed_hmac_valid_v0(http.request.uri, \"${var.environment}-hls-hmac-secret-key-v1\", 3600, http.request.timestamp.sec, 8))"
    description = "Block unsigned or expired HLS manifest/segment requests"
    enabled     = true
  }
}

# ==============================================================================
# 3. CLOUDFLARE STREAM LIVE INPUTS PROVISIONING POLICY
# ==============================================================================
resource "cloudflare_record" "live_ingest_cname" {
  zone_id = var.cloudflare_zone_id
  name    = "live-ingest"
  content = "live.cloudflare.com"
  type    = "CNAME"
  proxied = false
  ttl     = 300
}

# ==============================================================================
# 4. POSTGRESQL 18 HIGH-AVAILABILITY CLUSTER (PRIMARY + READ REPLICA + PITR)
# ==============================================================================
resource "aws_db_instance" "postgres_primary" {
  identifier                  = "gvp-${var.environment}-pg18-primary"
  engine                      = "postgres"
  engine_version              = "18.0"
  instance_class              = "db.r7g.2xlarge"
  allocated_storage           = 500
  max_allocated_storage       = 4000
  storage_type                = "gp3"
  storage_encrypted           = true
  multi_az                    = true
  db_name                     = "gvp_core"
  username                    = "gvp_admin"
  manage_master_user_password = true

  backup_retention_period   = 35
  backup_window             = "18:00-19:00"
  copy_tags_to_snapshot     = true
  deletion_protection       = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "gvp-${var.environment}-pg18-final-snapshot"

  performance_insights_enabled = true
  tags = {
    Project     = "GlobalVideoPlatform"
    Environment = var.environment
    Component   = "PostgreSQL-18-HA"
  }
}

resource "aws_db_instance" "postgres_read_replica" {
  identifier             = "gvp-${var.environment}-pg18-replica-1"
  replicate_source_db    = aws_db_instance.postgres_primary.identifier
  instance_class         = "db.r7g.2xlarge"
  storage_encrypted      = true
  publicly_accessible    = false
  skip_final_snapshot    = true
  performance_insights_enabled = true
}

# ==============================================================================
# 5. VALKEY 8 CLUSTER (CACHE, SESSION, RATE-LIMIT, REALTIME PUB/SUB)
# ==============================================================================
resource "aws_elasticache_replication_group" "valkey_cluster" {
  replication_group_id       = "gvp-${var.environment}-valkey"
  description                = "Valkey 8 HA cluster for GVP cache, rate limiting, and chat fanout"
  engine                     = "valkey"
  engine_version             = "8.0"
  node_type                  = "cache.r7g.xlarge"
  num_cache_clusters         = 3
  port                       = 6379
  automatic_failover_enabled = true
  multi_az_enabled           = true
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
}

# ==============================================================================
# 6. OPENSEARCH DOMAIN (VIDEO, CHANNEL & TRANSCRIPT SEARCH)
# ==============================================================================
resource "aws_opensearch_domain" "search_cluster" {
  domain_name    = "gvp-${var.environment}-search"
  engine_version = "OpenSearch_2.17"

  cluster_config {
    instance_type            = "r6g.large.search"
    instance_count           = 3
    zone_awareness_enabled   = true
    dedicated_master_enabled = true
    dedicated_master_type    = "m6g.large.search"
    dedicated_master_count   = 3
  }

  ebs_options {
    ebs_enabled = true
    volume_type = "gp3"
    volume_size = 250
  }

  encrypt_at_rest {
    enabled = true
  }

  node_to_node_encryption {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }
}

# ==============================================================================
# 7. KUBERNETES NODE POOLS (GENERAL API/WEB + DEDICATED FFMPEG TRANSCODING)
# ==============================================================================
resource "aws_eks_node_group" "general_workloads" {
  cluster_name    = "gvp-${var.environment}-eks"
  node_group_name = "gvp-general-pool"
  node_role_arn   = "arn:aws:iam::123456789012:role/gvp-eks-node-role"
  subnet_ids      = ["subnet-prod-1a", "subnet-prod-1b", "subnet-prod-1c"]
  instance_types  = ["m7i.xlarge"]

  scaling_config {
    desired_size = 6
    min_size     = 4
    max_size     = 24
  }

  labels = {
    "gvp.io/node-pool" = "general"
  }
}

resource "aws_eks_node_group" "media_transcode_pool" {
  cluster_name    = "gvp-${var.environment}-eks"
  node_group_name = "gvp-media-transcode-pool"
  node_role_arn   = "arn:aws:iam::123456789012:role/gvp-eks-node-role"
  subnet_ids      = ["subnet-prod-1a", "subnet-prod-1b", "subnet-prod-1c"]
  capacity_type   = "SPOT"
  instance_types  = ["c7i.4xlarge", "c6i.4xlarge"]

  scaling_config {
    desired_size = 3
    min_size     = 2
    max_size     = 32
  }

  labels = {
    "gvp.io/node-pool" = "media-transcode"
  }
}
