package store

import (
	"sync"
	"time"
)

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Username         string    `json:"username"`
	Role             string    `json:"role"`
	SubscriptionTier string    `json:"subscription_tier"`
	IsVerified       bool      `json:"is_verified"`
	IsActive         bool      `json:"is_active"`
	DisplayName      string    `json:"display_name"`
	Bio              string    `json:"bio"`
	AvatarURL        string    `json:"avatar_url"`
	BannerURL        string    `json:"banner_url"`
	CreatedAt        time.Time `json:"created_at"`
}

type Channel struct {
	ID                  string    `json:"id"`
	OwnerID             string    `json:"owner_id"`
	Handle              string    `json:"handle"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	AvatarURL           string    `json:"avatar_url"`
	BannerURL           string    `json:"banner_url"`
	SubscriberCount     int64     `json:"subscriber_count"`
	VideoCount          int       `json:"video_count"`
	IsVerified          bool      `json:"is_verified"`
	MonetizationEnabled bool      `json:"monetization_enabled"`
	CreatedAt           time.Time `json:"created_at"`
}

type VideoChapter struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	StartTimeSeconds int    `json:"start_time_seconds"`
	EndTimeSeconds   int    `json:"end_time_seconds"`
}

type VideoCaption struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Label    string `json:"label"`
	VttURL   string `json:"vtt_url"`
}

type VideoSource struct {
	Quality     string `json:"quality"`
	ManifestURL string `json:"manifest_url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Bitrate     int64  `json:"bitrate"`
}

type Video struct {
	ID              string         `json:"id"`
	ChannelID       string         `json:"channel_id"`
	ChannelName     string         `json:"channel_name,omitempty"`
	ChannelHandle   string         `json:"channel_handle,omitempty"`
	ChannelAvatar   string         `json:"channel_avatar_url,omitempty"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	VideoType       string         `json:"video_type"` // standard, shorts, live_vod
	Visibility      string         `json:"visibility"` // public, unlisted, private, members_only
	Status          string         `json:"status"`     // CREATED, PROCESSING, READY, FAILED
	DurationSeconds int            `json:"duration_seconds"`
	ThumbnailURL    string         `json:"thumbnail_url"`
	HlsMasterURL    string         `json:"hls_master_url"`
	ViewCount       int64          `json:"view_count"`
	LikeCount       int64          `json:"like_count"`
	DislikeCount    int64          `json:"dislike_count"`
	CommentCount    int64          `json:"comment_count"`
	Tags            []string       `json:"tags"`
	Category        string         `json:"category"`
	Sources         []VideoSource  `json:"sources"`
	Chapters        []VideoChapter `json:"chapters"`
	Captions        []VideoCaption `json:"captions"`
	PublishedAt     time.Time      `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Comment struct {
	ID           string    `json:"id"`
	VideoID      string    `json:"video_id"`
	UserID       string    `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	Content      string    `json:"content"`
	LikeCount    int       `json:"like_count"`
	ReplyCount   int       `json:"reply_count"`
	IsPinned     bool      `json:"is_pinned"`
	IsHearted    bool      `json:"is_hearted"`
	CreatedAt    time.Time `json:"created_at"`
}

type LiveStream struct {
	ID             string    `json:"id"`
	ChannelID      string    `json:"channel_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	Status         string    `json:"status"` // idle, live, ended
	StreamKey      string    `json:"stream_key"`
	RtmpsIngestURL string    `json:"rtmps_ingest_url"`
	SrtIngestURL   string    `json:"srt_ingest_url"`
	PlaybackHlsURL string    `json:"playback_hls_url"`
	ViewerCount    int       `json:"viewer_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type AdCampaign struct {
	ID               string    `json:"id"`
	AdvertiserID     string    `json:"advertiser_id"`
	Name             string    `json:"name"`
	Format           string    `json:"format"`
	Status           string    `json:"status"`
	TotalBudgetCents int64     `json:"total_budget_cents"`
	DailyBudgetCents int64     `json:"daily_budget_cents"`
	SpentCents       int64     `json:"spent_cents"`
	Impressions      int64     `json:"impressions"`
	Clicks           int64     `json:"clicks"`
	CreatedAt        time.Time `json:"created_at"`
}

type CopyrightClaim struct {
	ID                 string    `json:"id"`
	ClaimedVideoID     string    `json:"claimed_video_id"`
	OwnerName          string    `json:"owner_name"`
	ClaimAction        string    `json:"claim_action"`
	Status             string    `json:"status"`
	MatchStartSeconds  int       `json:"match_start_seconds"`
	MatchEndSeconds    int       `json:"match_end_seconds"`
	CreatedAt          time.Time `json:"created_at"`
}

type ModerationCase struct {
	ID             string    `json:"id"`
	TargetType     string    `json:"target_type"`
	TargetID       string    `json:"target_id"`
	Reason         string    `json:"reason"`
	PolicyCategory string    `json:"policy_category"`
	Status         string    `json:"status"`
	ActionTaken    string    `json:"action_taken,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PlaylistItem struct {
	VideoID  string    `json:"video_id"`
	Position int       `json:"position"`
	AddedAt  time.Time `json:"added_at"`
}

type Playlist struct {
	ID          string         `json:"id"`
	OwnerID     string         `json:"owner_id"`
	ChannelID   string         `json:"channel_id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Visibility  string         `json:"visibility"` // public, unlisted, private
	Items       []PlaylistItem `json:"items"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type LedgerEntry struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	AccountType   string    `json:"account_type"` // user_wallet, platform_escrow, creator_payable, platform_revenue
	Direction     string    `json:"direction"`    // DEBIT, CREDIT
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type PaymentTransaction struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	Provider       string        `json:"provider"` // stripe, toss
	IdempotencyKey string        `json:"idempotency_key"`
	Type           string        `json:"type"` // membership, super_chat, ad_topup, premium_subscription
	AmountCents    int64         `json:"amount_cents"`
	Currency       string        `json:"currency"`
	Status         string        `json:"status"` // pending, succeeded, failed, refunded
	CheckoutURL    string        `json:"checkout_url,omitempty"`
	LedgerEntries  []LedgerEntry `json:"ledger_entries,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type MemoryStore struct {
	mu              sync.RWMutex
	Users           map[string]*User
	Channels        map[string]*Channel
	Videos          map[string]*Video
	Comments        map[string][]*Comment
	Subscriptions   map[string]map[string]bool // userID -> set of channelIDs
	Likes           map[string]map[string]bool // userID -> set of videoIDs
	LiveStreams     map[string]*LiveStream
	Campaigns       map[string]*AdCampaign
	Claims          map[string]*CopyrightClaim
	Moderation      map[string]*ModerationCase
	Playlists       map[string]*Playlist
	Transactions    map[string]*PaymentTransaction
	IdempotencyKeys map[string]string // idempotencyKey -> transactionID
	Ledger          []LedgerEntry
	AuditLogs       []map[string]interface{}
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		Users:           make(map[string]*User),
		Channels:        make(map[string]*Channel),
		Videos:          make(map[string]*Video),
		Comments:        make(map[string][]*Comment),
		Subscriptions:   make(map[string]map[string]bool),
		Likes:           make(map[string]map[string]bool),
		LiveStreams:     make(map[string]*LiveStream),
		Campaigns:       make(map[string]*AdCampaign),
		Claims:          make(map[string]*CopyrightClaim),
		Moderation:      make(map[string]*ModerationCase),
		Playlists:       make(map[string]*Playlist),
		Transactions:    make(map[string]*PaymentTransaction),
		IdempotencyKeys: make(map[string]string),
		Ledger:          make([]LedgerEntry, 0),
		AuditLogs:       make([]map[string]interface{}, 0),
	}
	s.seedDefaults()
	return s
}

func (s *MemoryStore) seedDefaults() {
	now := time.Now()

	adminUser := &User{
		ID:               "11111111-1111-1111-1111-111111111111",
		Email:            "admin@brand.com",
		Username:         "admin_master",
		Role:             "super_admin",
		SubscriptionTier: "premium",
		IsVerified:       true,
		IsActive:         true,
		DisplayName:      "System Administrator",
		Bio:              "Global Video Platform Root Administration",
		AvatarURL:        "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150",
		CreatedAt:        now,
	}
	s.Users[adminUser.ID] = adminUser

	creatorUser := &User{
		ID:               "22222222-2222-2222-2222-222222222222",
		Email:            "creator1@brand.com",
		Username:         "tech_nexus",
		Role:             "creator",
		SubscriptionTier: "premium",
		IsVerified:       true,
		IsActive:         true,
		DisplayName:      "TechNexus Lab",
		Bio:              "Future Technology, AI & Next-Gen Engineering",
		AvatarURL:        "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=150",
		CreatedAt:        now,
	}
	s.Users[creatorUser.ID] = creatorUser

	channel1 := &Channel{
		ID:                  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		OwnerID:             creatorUser.ID,
		Handle:              "@technexus",
		Name:                "TechNexus",
		Description:         "Deep dives into next-generation systems, AI architectures, and developer ecosystems.",
		AvatarURL:           "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=200",
		BannerURL:           "https://images.unsplash.com/photo-1518770660439-4636190af475?w=1200",
		SubscriberCount:     482000,
		VideoCount:          142,
		IsVerified:          true,
		MonetizationEnabled: true,
		CreatedAt:           now,
	}
	s.Channels[channel1.ID] = channel1

	video1 := &Video{
		ID:              "00000000-0000-0000-0000-000000000001",
		ChannelID:       channel1.ID,
		ChannelName:     channel1.Name,
		ChannelHandle:   channel1.Handle,
		ChannelAvatar:   channel1.AvatarURL,
		Title:           "Building a Distributed Video Engine at Scale with Go & Cloudflare R2",
		Description:     "In this in-depth masterclass, we explore how to construct an enterprise video architecture with zero single-points-of-failure.",
		VideoType:       "standard",
		Visibility:      "public",
		Status:          "READY",
		DurationSeconds: 754,
		ThumbnailURL:    "https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=800",
		HlsMasterURL:    "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
		ViewCount:       125430,
		LikeCount:       8920,
		DislikeCount:    45,
		CommentCount:    342,
		Tags:            []string{"go", "video", "streaming", "architecture", "hls"},
		Category:        "Technology",
		Sources: []VideoSource{
			{Quality: "1080p", ManifestURL: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8", Width: 1920, Height: 1080, Bitrate: 4500000},
			{Quality: "720p", ManifestURL: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8", Width: 1280, Height: 720, Bitrate: 2500000},
			{Quality: "480p", ManifestURL: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8", Width: 854, Height: 480, Bitrate: 1200000},
		},
		Chapters: []VideoChapter{
			{ID: "c1", Title: "System Overview", StartTimeSeconds: 0, EndTimeSeconds: 120},
			{ID: "c2", Title: "Direct Multipart Uploads", StartTimeSeconds: 120, EndTimeSeconds: 380},
			{ID: "c3", Title: "FFmpeg Transcoding Pipeline", StartTimeSeconds: 380, EndTimeSeconds: 580},
			{ID: "c4", Title: "Adaptive Bitrate with Shaka", StartTimeSeconds: 580, EndTimeSeconds: 754},
		},
		Captions: []VideoCaption{
			{ID: "cap1", Language: "ko", Label: "한국어", VttURL: "https://example.com/ko.vtt"},
			{ID: "cap2", Language: "en", Label: "English", VttURL: "https://example.com/en.vtt"},
		},
		PublishedAt: now.Add(-72 * time.Hour),
		CreatedAt:   now.Add(-72 * time.Hour),
	}
	s.Videos[video1.ID] = video1

	videoShort := &Video{
		ID:              "00000000-0000-0000-0000-000000000004",
		ChannelID:       channel1.ID,
		ChannelName:     channel1.Name,
		ChannelHandle:   channel1.Handle,
		ChannelAvatar:   channel1.AvatarURL,
		Title:           "3 Hidden Go 1.27 Tricks You Probably Do Not Know #shorts",
		Description:     "Lightning quick demonstration of brand new Go runtime compiler optimizations.",
		VideoType:       "shorts",
		Visibility:      "public",
		Status:          "READY",
		DurationSeconds: 42,
		ThumbnailURL:    "https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=600",
		HlsMasterURL:    "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
		ViewCount:       358900,
		LikeCount:       24100,
		DislikeCount:    120,
		CommentCount:    520,
		Tags:            []string{"golang", "shorts", "coding"},
		Category:        "Technology",
		PublishedAt:     now.Add(-24 * time.Hour),
		CreatedAt:       now.Add(-24 * time.Hour),
	}
	s.Videos[videoShort.ID] = videoShort

	s.Comments[video1.ID] = []*Comment{
		{
			ID:          "comm-1",
			VideoID:     video1.ID,
			UserID:      "44444444-4444-4444-4444-444444444444",
			DisplayName: "Sarah Kim",
			AvatarURL:   "https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=150",
			Content:     "This is one of the best explanations of adaptive HLS streaming I have seen! The Go modular monolith design is crystal clear.",
			LikeCount:   42,
			ReplyCount:  1,
			IsPinned:    true,
			CreatedAt:   now.Add(-48 * time.Hour),
		},
	}
}

