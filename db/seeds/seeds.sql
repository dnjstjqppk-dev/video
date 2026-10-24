-- Seed Initial Platform Data for Development & Testing

-- 1. Users
INSERT INTO users (id, email, username, role, subscription_tier, is_verified, is_active) VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin@brand.com', 'admin_master', 'super_admin', 'premium', TRUE, TRUE),
    ('22222222-2222-2222-2222-222222222222', 'creator1@brand.com', 'tech_nexus', 'creator', 'premium', TRUE, TRUE),
    ('33333333-3333-3333-3333-333333333333', 'creator2@brand.com', 'culinary_craft', 'creator', 'free', TRUE, TRUE),
    ('44444444-4444-4444-4444-444444444444', 'viewer1@brand.com', 'sarah_k', 'registered_user', 'premium', FALSE, TRUE),
    ('55555555-5555-5555-5555-555555555555', 'partner@brand.com', 'acme_marketing', 'advertiser', 'free', TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 2. User Profiles
INSERT INTO user_profiles (user_id, display_name, bio, avatar_url, language) VALUES
    ('11111111-1111-1111-1111-111111111111', 'System Administrator', 'Global Video Platform Root Administration', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150', 'ko'),
    ('22222222-2222-2222-2222-222222222222', 'TechNexus Lab', 'Future Technology, AI & Next-Gen Engineering', 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=150', 'ko'),
    ('33333333-3333-3333-3333-333333333333', 'Culinary Craft', 'Artisanal recipes and gastronomy documentary', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=150', 'ko'),
    ('44444444-4444-4444-4444-444444444444', 'Sarah Kim', 'Video creator and tech enthusiast', 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=150', 'ko'),
    ('55555555-5555-5555-5555-555555555555', 'Acme Media Corp', 'Global Digital Performance Agency', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', 'en')
ON CONFLICT (user_id) DO NOTHING;

-- 3. Channels
INSERT INTO channels (id, owner_id, handle, name, description, avatar_url, banner_url, subscriber_count, video_count, is_verified, monetization_enabled) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222', '@technexus', 'TechNexus', 'Deep dives into next-generation systems, AI architectures, and developer ecosystems.', 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=200', 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=1200', 482000, 142, TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', '@culinarycraft', 'Culinary Craft', 'Mastering the art of cooking with precise culinary science.', 'https://images.unsplash.com/photo-1556910103-1c02745aae4d?w=200', 'https://images.unsplash.com/photo-1504674900247-0877df9cc836?w=1200', 129000, 85, TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 4. Sample Videos with HLS adaptive streams (Test stream URLs)
INSERT INTO videos (id, channel_id, title, description, video_type, visibility, status, duration_seconds, thumbnail_url, hls_master_url, view_count, like_count, comment_count, tags, category, published_at) VALUES
    ('00000000-0000-0000-0000-000000000001', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Building a Distributed Video Engine at Scale with Go & Cloudflare R2', 'In this in-depth masterclass, we explore how to construct an enterprise video architecture with zero single-points-of-failure.', 'standard', 'public', 'READY', 754, 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=800', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 125430, 8920, 342, ARRAY['go', 'video', 'streaming', 'architecture', 'hls'], 'Technology', NOW() - INTERVAL '3 days'),
    ('00000000-0000-0000-0000-000000000002', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Vue 3 + Nuxt 4 Architecture Guide for High Traffic Apps', 'Everything you need to know about Nuxt 4 server components, hybrid rendering, and high-performance Pinia state management.', 'standard', 'public', 'READY', 1120, 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 84210, 5630, 219, ARRAY['vue', 'nuxt', 'typescript', 'frontend'], 'Technology', NOW() - INTERVAL '7 days'),
    ('00000000-0000-0000-0000-000000000003', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'The 72-Hour Sourdough Bread Masterclass', 'Achieve the ultimate open crumb and crispy blistered crust with high hydration sourdough fermentation.', 'standard', 'public', 'READY', 890, 'https://images.unsplash.com/photo-1509440159596-0249088772ff?w=800', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 214500, 14200, 680, ARRAY['baking', 'sourdough', 'cooking', 'artisan'], 'Food', NOW() - INTERVAL '10 days'),
    -- Shorts
    ('00000000-0000-0000-0000-000000000004', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '3 Hidden Go 1.27 Tricks You Probably Do Not Know #shorts', 'Lightning quick demonstration of brand new Go 1.27 runtime compiler optimizations.', 'shorts', 'public', 'READY', 42, 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=600', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 358900, 24100, 520, ARRAY['golang', 'shorts', 'coding'], 'Technology', NOW() - INTERVAL '1 day'),
    ('00000000-0000-0000-0000-000000000005', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Perfect French Omelette in 60 Seconds #shorts', 'No browning, silky smooth interior, and pure butter richness.', 'shorts', 'public', 'READY', 58, 'https://images.unsplash.com/photo-1525351484163-7529414344d8?w=600', 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8', 492000, 39400, 890, ARRAY['cooking', 'egg', 'shorts', 'french'], 'Food', NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- 5. Chapters
INSERT INTO chapters (video_id, title, start_time_seconds, end_time_seconds) VALUES
    ('00000000-0000-0000-0000-000000000001', 'System Overview & Problem Statement', 0, 120),
    ('00000000-0000-0000-0000-000000000001', 'Direct Multipart Uploads to S3/R2', 120, 380),
    ('00000000-0000-0000-0000-000000000001', 'FFmpeg Transcoding Pipeline & HLS Variants', 380, 580),
    ('00000000-0000-0000-0000-000000000001', 'Adaptive Bitrate Streaming with Shaka Player', 580, 754)
ON CONFLICT DO NOTHING;

-- 6. Comments
INSERT INTO comments (id, video_id, user_id, content, like_count) VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccc0001', '00000000-0000-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'This is one of the best explanations of adaptive HLS streaming I have seen! The Go modular monolith design is crystal clear.', 42),
    ('cccccccc-cccc-cccc-cccc-cccccccc0002', '00000000-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'Thank you! In the next video we will connect the Cloudflare Stream live pipeline.', 18)
ON CONFLICT (id) DO NOTHING;

