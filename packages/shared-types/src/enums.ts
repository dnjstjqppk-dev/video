// ==========================================
// Global Video Platform (GVP) Shared Enums
// ==========================================

export enum UserRole {
  VISITOR = 'visitor',
  REGISTERED_USER = 'registered_user',
  CREATOR = 'creator',
  CHANNEL_MANAGER = 'channel_manager',
  CHANNEL_EDITOR = 'channel_editor',
  CHANNEL_ANALYST = 'channel_analyst',
  ADVERTISER = 'advertiser',
  MCN_MANAGER = 'mcn_manager',
  RIGHTS_OWNER = 'rights_owner',
  PARTNER_ADMIN = 'partner_admin',
  MODERATOR = 'moderator',
  SUPPORT_AGENT = 'support_agent',
  FINANCE_ADMIN = 'finance_admin',
  PLATFORM_ADMIN = 'platform_admin',
  SUPER_ADMIN = 'super_admin'
}

export enum VideoVisibility {
  PUBLIC = 'public',
  UNLISTED = 'unlisted',
  PRIVATE = 'private',
  MEMBERS_ONLY = 'members_only',
  SCHEDULED = 'scheduled'
}

export enum VideoProcessingStatus {
  CREATED = 'CREATED',
  UPLOADING = 'UPLOADING',
  UPLOADED = 'UPLOADED',
  VALIDATING = 'VALIDATING',
  QUEUED = 'QUEUED',
  PROCESSING = 'PROCESSING',
  READY = 'READY',
  FAILED = 'FAILED',
  BLOCKED = 'BLOCKED',
  DELETED = 'DELETED'
}

export enum VideoQuality {
  P240 = '240p',
  P360 = '360p',
  P480 = '480p',
  P720 = '720p',
  P1080 = '1080p',
  P1440 = '1440p',
  P2160 = '2160p'
}

export enum VideoType {
  STANDARD = 'standard',
  SHORTS = 'shorts',
  LIVE_VOD = 'live_vod'
}

export enum LiveStreamStatus {
  IDLE = 'idle',
  PREPARING = 'preparing',
  LIVE = 'live',
  DISCONNECTED = 'disconnected',
  ENDED = 'ended'
}

export enum AdFormat {
  PREROLL = 'preroll',
  MIDROLL = 'midroll',
  FEED = 'feed',
  SHORTS = 'shorts',
  DISPLAY = 'display'
}

export enum CampaignStatus {
  DRAFT = 'draft',
  PENDING_REVIEW = 'pending_review',
  ACTIVE = 'active',
  PAUSED = 'paused',
  COMPLETED = 'completed',
  REJECTED = 'rejected'
}

export enum ModerationStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  APPEALED = 'appealed',
  ESCALATED = 'escalated'
}

export enum CopyrightClaimAction {
  BLOCK = 'block',
  MONETIZE = 'monetize',
  TRACK = 'track'
}

export enum CopyrightClaimStatus {
  PENDING = 'pending',
  ACTIVE = 'active',
  DISPUTED = 'disputed',
  RESOLVED = 'resolved',
  WITHDRAWN = 'withdrawn'
}

export enum LedgerAccountType {
  ASSET = 'asset',
  LIABILITY = 'liability',
  EQUITY = 'equity',
  REVENUE = 'revenue',
  EXPENSE = 'expense'
}

export enum PayoutStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  PROCESSING = 'processing',
  PAID = 'paid',
  FAILED = 'failed',
  ON_HOLD = 'on_hold'
}

export enum SubscriptionTier {
  FREE = 'free',
  PREMIUM = 'premium',
  PREMIUM_FAMILY = 'premium_family'
}

