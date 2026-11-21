package store

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// pgTestStore connects to TEST_DATABASE_URL or skips. Run a migrated
// throwaway Postgres (db/migrations/00001.. applied) to exercise this suite.
func pgTestStore(t *testing.T) *PostgresStore {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset; skipping Postgres round-trip suite")
	}
	pg, err := NewPostgresStore(dsn)
	if err != nil {
		t.Fatalf("connect test postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Close() })
	return pg
}

func pgCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// pgCleanup deletes exactly the rows a round-trip test created (caller
// passes {table, column, id} in FK-safe order) so the throwaway database
// never leaks fixtures into hydrated serving.
func pgCleanup(t *testing.T, pg *PostgresStore, rows ...[3]string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, r := range rows {
			if _, err := pg.DB.ExecContext(ctx, "DELETE FROM "+r[0]+" WHERE "+r[1]+" = $1", r[2]); err != nil {
				t.Errorf("cleanup %s.%s=%s: %v", r[0], r[1], r[2], err)
			}
		}
	})
}

// P1-13: job status mapping across the memory/persistence boundary.
func TestMediaJobStatusMapping(t *testing.T) {
	cases := map[string][2]string{
		"QUEUED":     {"QUEUED", "QUEUED"},
		"PROCESSING": {"RUNNING", "QUEUED"},
		"READY":      {"SUCCEEDED", "READY"},
		"FAILED":     {"FAILED", "FAILED"},
	}
	for memory, want := range cases {
		if got := memoryToPGJobStatus(memory); got != want[0] {
			t.Fatalf("memoryToPG(%s) = %s, want %s", memory, got, want[0])
		}
	}
	pgCases := map[string]string{"QUEUED": "QUEUED", "RUNNING": "QUEUED", "SUCCEEDED": "READY", "FAILED": "FAILED"}
	for pg, want := range pgCases {
		if got := pgToMemoryJobStatus(pg); got != want {
			t.Fatalf("pgToMemory(%s) = %s, want %s", pg, got, want)
		}
	}
}

// P1-13: media jobs persist and rehydrate (RUNNING returns as QUEUED).
func TestPostgresMediaJobRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	vidID := uuid.NewString()
	ownerID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: ownerID, Email: "job-" + ownerID[:8] + "@test.local", Username: "job" + ownerID[:8],
		Role: "creator", SubscriptionTier: "free", IsActive: true, DisplayName: "Job",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser: %v", err)
	}
	chID := uuid.NewString()
	if err := pg.SaveChannel(ctx, &Channel{
		ID: chID, OwnerID: ownerID, Handle: "@job" + chID[:8], Name: "Job Ch",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}
	if err := pg.SaveVideo(ctx, &Video{
		ID: vidID, ChannelID: chID, Title: "Job Vid", Visibility: "public",
		Status: "READY", CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveVideo: %v", err)
	}
	job := &MediaJob{
		ID: uuid.NewString(), VideoID: vidID, InputURI: "file:///tmp/x.mp4",
		Status: "PROCESSING", ProgressPct: 10, CreatedAt: time.Now().UTC(),
	}
	if err := pg.SaveMediaJob(ctx, job); err != nil {
		t.Fatalf("SaveMediaJob: %v", err)
	}
	jobs, err := pg.LoadMediaJobs(ctx)
	if err != nil {
		t.Fatalf("LoadMediaJobs: %v", err)
	}
	found := false
	for _, j := range jobs {
		if j.ID == job.ID {
			found = true
			if j.Status != "QUEUED" {
				t.Fatalf("RUNNING must rehydrate as QUEUED, got %s", j.Status)
			}
			if j.InputURI != "file:///tmp/x.mp4" {
				t.Fatalf("source_ref lost: %q", j.InputURI)
			}
		}
	}
	if !found {
		t.Fatal("saved job missing after load")
	}
	pgCleanup(t, pg,
		[3]string{"media_jobs", "id", job.ID},
		[3]string{"videos", "id", vidID},
		[3]string{"channels", "id", chID},
		[3]string{"users", "id", ownerID},
	)
}

// P0-1: users persist to the real migration schema (users + user_profiles)
// and rehydrate on a fresh store (restart durability).
func TestPostgresUserRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	id := uuid.NewString()
	u := &User{
		ID: id, Email: "pg-" + id[:8] + "@test.local", Username: "pg-user" + id[:8],
		Role: "creator", SubscriptionTier: "free", IsVerified: true, IsActive: true,
		DisplayName: "PG Test", Bio: "round trip", AvatarURL: "https://test.local/a.png",
		CreatedAt: time.Now().UTC().Round(time.Millisecond),
	}
	if err := pg.SaveUser(ctx, u); err != nil {
		t.Fatalf("SaveUser against migrated schema: %v", err)
	}
	got, err := pg.LoadUser(ctx, id)
	if err != nil {
		t.Fatalf("LoadUser: %v", err)
	}
	if got.Email != u.Email || got.DisplayName != u.DisplayName || got.Role != u.Role {
		t.Fatalf("user mismatch after round trip: %+v", got)
	}

	// Restart durability: a fresh memory store hydrates the persisted row.
	mem := NewMemoryStoreWithDB(pg)
	secondCtx, secondCancel := pgCtx()
	defer secondCancel()
	if err := mem.Hydrate(secondCtx); err != nil {
		t.Fatalf("Hydrate: %v", err)
	}
	hydrated, ok := mem.Users[id]
	if !ok || hydrated.Email != u.Email || hydrated.DisplayName != u.DisplayName {
		t.Fatalf("hydrated user missing or wrong: %+v", hydrated)
	}
	pgCleanup(t, pg, [3]string{"users", "id", id})
}

// P0-1: channels and videos persist with FK integrity and rehydrate.
func TestPostgresChannelVideoRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	ownerID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: ownerID, Email: "own-" + ownerID[:8] + "@test.local", Username: "own" + ownerID[:8],
		Role: "creator", SubscriptionTier: "free", IsActive: true, DisplayName: "Owner",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser owner: %v", err)
	}
	chID := uuid.NewString()
	if err := pg.SaveChannel(ctx, &Channel{
		ID: chID, OwnerID: ownerID, Handle: "@pg" + chID[:8], Name: "PG Channel",
		SubscriberCount: 10, VideoCount: 1, CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}
	vidID := uuid.NewString()
	if err := pg.SaveVideo(ctx, &Video{
		ID: vidID, ChannelID: chID, Title: "PG Video", Description: "d",
		VideoType: "standard", Visibility: "public", Status: "READY",
		DurationSeconds: 60, ThumbnailURL: "https://test.local/t.png",
		HlsMasterURL: "https://test.local/m.m3u8", ViewCount: 5, LikeCount: 2,
		Tags: []string{"a", "b"}, Category: "tech", CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveVideo against migrated schema: %v", err)
	}

	mem := NewMemoryStoreWithDB(pg)
	secondCtx, secondCancel := pgCtx()
	defer secondCancel()
	if err := mem.Hydrate(secondCtx); err != nil {
		t.Fatalf("Hydrate: %v", err)
	}
	v, ok := mem.Videos[vidID]
	if !ok || v.Title != "PG Video" || len(v.Tags) != 2 || v.ViewCount != 5 {
		t.Fatalf("hydrated video missing or wrong: %+v", v)
	}
	if _, ok := mem.Channels[chID]; !ok {
		t.Fatal("hydrated channel missing")
	}
	pgCleanup(t, pg,
		[3]string{"videos", "id", vidID},
		[3]string{"channels", "id", chID},
		[3]string{"users", "id", ownerID},
	)
}

// P0-1: payment transactions and ledger entries persist atomically and rehydrate.
func TestPostgresTransactionRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	userID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: userID, Email: "pay-" + userID[:8] + "@test.local", Username: "pay" + userID[:8],
		Role: "creator", SubscriptionTier: "free", IsActive: true, DisplayName: "Pay",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser payer: %v", err)
	}
	txID := uuid.NewString()
	tx := &PaymentTransaction{
		ID: txID, UserID: userID, Provider: "stripe", IdempotencyKey: "idem-" + txID,
		Type: "membership", AmountCents: 5000, Currency: "KRW", Status: "succeeded",
		CheckoutURL: "https://checkout.test/s/1", CreatedAt: time.Now().UTC(),
		LedgerEntries: []LedgerEntry{
			{ID: uuid.NewString(), TransactionID: txID, AccountType: "platform_escrow", Direction: "DEBIT", AmountCents: 5000, Currency: "KRW", CreatedAt: time.Now().UTC()},
			{ID: uuid.NewString(), TransactionID: txID, AccountType: "creator_payable", Direction: "CREDIT", AmountCents: 3500, Currency: "KRW", CreatedAt: time.Now().UTC()},
			{ID: uuid.NewString(), TransactionID: txID, AccountType: "platform_revenue", Direction: "CREDIT", AmountCents: 1500, Currency: "KRW", CreatedAt: time.Now().UTC()},
		},
	}
	if err := pg.SaveTransaction(ctx, tx); err != nil {
		t.Fatalf("SaveTransaction against migrated schema: %v", err)
	}

	mem := NewMemoryStoreWithDB(pg)
	secondCtx, secondCancel := pgCtx()
	defer secondCancel()
	if err := mem.Hydrate(secondCtx); err != nil {
		t.Fatalf("Hydrate: %v", err)
	}
	got, ok := mem.Transactions[txID]
	if !ok || got.Status != "succeeded" || len(got.LedgerEntries) != 3 {
		t.Fatalf("hydrated transaction missing or wrong: %+v", got)
	}
	var debits, credits int64
	for _, e := range got.LedgerEntries {
		if e.Direction == "DEBIT" {
			debits += e.AmountCents
		} else {
			credits += e.AmountCents
		}
	}
	if debits != credits || debits != 5000 {
		t.Fatalf("ledger must balance at 5000, got debit=%d credit=%d", debits, credits)
	}
	pgCleanup(t, pg,
		[3]string{"ledger_entries", "transaction_id", txID},
		[3]string{"payment_transactions", "id", txID},
		[3]string{"users", "id", userID},
	)
}

// View counting must bump only the counter: a full-row rewrite would freeze
// seed content in PostgreSQL and shadow later seed updates on hydrate.
func TestPostgresBumpVideoViewCount_TouchesCounterOnly(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	ownerID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: ownerID, Email: "vc-" + ownerID[:8] + "@test.local", Username: "vc" + ownerID[:8],
		Role: "creator", SubscriptionTier: "free", IsActive: true, DisplayName: "VC",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser: %v", err)
	}
	chID := uuid.NewString()
	if err := pg.SaveChannel(ctx, &Channel{
		ID: chID, OwnerID: ownerID, Handle: "@vc" + chID[:8], Name: "VC Ch",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}
	vidID := uuid.NewString()
	if err := pg.SaveVideo(ctx, &Video{
		ID: vidID, ChannelID: chID, Title: "Counter Probe", Visibility: "public",
		Status: "READY", CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveVideo: %v", err)
	}
	pgCleanup(t, pg,
		[3]string{"videos", "id", vidID},
		[3]string{"channels", "id", chID},
		[3]string{"users", "id", ownerID},
	)
	if err := pg.BumpVideoViewCount(ctx, vidID); err != nil {
		t.Fatalf("bump 1: %v", err)
	}
	if err := pg.BumpVideoViewCount(ctx, vidID); err != nil {
		t.Fatalf("bump 2: %v", err)
	}
	videos, err := pg.LoadVideos(ctx)
	if err != nil {
		t.Fatalf("LoadVideos: %v", err)
	}
	for _, v := range videos {
		if v.ID != vidID {
			continue
		}
		if v.ViewCount != 2 {
			t.Fatalf("view_count = %d, want 2", v.ViewCount)
		}
		if v.Title != "Counter Probe" {
			t.Fatalf("title rewritten by counter bump: %q", v.Title)
		}
		return
	}
	t.Fatal("probe video missing after bump")
}

// AI home ground + DRM: the 00013/00014 columns and the series table
// persist and rehydrate, and the migrated schema verifies clean.
func TestPostgresAIVideoSeriesRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()
	if err := pg.VerifySchema(ctx); err != nil {
		t.Fatalf("VerifySchema: %v", err)
	}

	ownerID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: ownerID, Email: "ai-" + ownerID[:8] + "@test.local", Username: "ai" + ownerID[:8],
		Role: "creator", SubscriptionTier: "free", IsActive: true, DisplayName: "AI",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser: %v", err)
	}
	chID := uuid.NewString()
	if err := pg.SaveChannel(ctx, &Channel{
		ID: chID, OwnerID: ownerID, Handle: "@ai" + chID[:8], Name: "AI Ch",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}
	now := time.Now().UTC()
	sID := uuid.NewString()
	if err := pg.SaveSeries(ctx, &Series{
		ID: sID, ChannelID: chID, Title: "AI RT", Status: "ongoing",
		AiOriginal: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("SaveSeries: %v", err)
	}
	vidID := uuid.NewString()
	if err := pg.SaveVideo(ctx, &Video{
		ID: vidID, ChannelID: chID, Title: "AI RT E1", Visibility: "public",
		Status: "READY", AiGenerated: true, AiTool: "roundtrip", AiReviewStatus: "approved",
		SeriesID: sID, SeasonNo: 1, EpisodeNo: 1,
		DrmProtected: true, DrmKeyID: "kid-rt", DrmKeyHex: "a1b2c3d4e5f6",
		CreatedAt: now,
	}); err != nil {
		t.Fatalf("SaveVideo: %v", err)
	}

	videos, err := pg.LoadVideos(ctx)
	if err != nil {
		t.Fatalf("LoadVideos: %v", err)
	}
	var got *Video
	for _, v := range videos {
		if v.ID == vidID {
			got = v
			break
		}
	}
	if got == nil {
		t.Fatalf("saved AI video %s missing from LoadVideos (%d rows)", vidID, len(videos))
	}
	if !got.AiGenerated || got.AiTool != "roundtrip" || got.AiReviewStatus != "approved" {
		t.Fatalf("AI fields lost: %+v", got)
	}
	if got.SeriesID != sID || got.SeasonNo != 1 || got.EpisodeNo != 1 {
		t.Fatalf("series slot lost: %+v", got)
	}
	if !got.DrmProtected || got.DrmKeyID != "kid-rt" || got.DrmKeyHex != "a1b2c3d4e5f6" {
		t.Fatalf("DRM fields lost: %+v", got)
	}
	allSeries, err := pg.LoadSeries(ctx)
	if err != nil {
		t.Fatalf("LoadSeries: %v", err)
	}
	var gotS *Series
	for _, s := range allSeries {
		if s.ID == sID {
			gotS = s
			break
		}
	}
	if gotS == nil || !gotS.AiOriginal || gotS.Title != "AI RT" {
		t.Fatalf("saved series %s missing or wrong: %+v", sID, gotS)
	}
	pgCleanup(t, pg,
		[3]string{"videos", "id", vidID},
		[3]string{"series", "id", sID},
		[3]string{"channels", "id", chID},
		[3]string{"users", "id", ownerID},
	)
}

// P26: ad decisions persist with both UUID columns and legacy prefixed refs.
func TestPostgresDecisionRoundTrip(t *testing.T) {
	pg := pgTestStore(t)
	ctx, cancel := pgCtx()
	defer cancel()

	ownerID := uuid.NewString()
	if err := pg.SaveUser(ctx, &User{
		ID: ownerID, Email: "dec-" + ownerID[:8] + "@test.local", Username: "dec" + ownerID[:8],
		Role: "registered_user", SubscriptionTier: "free", IsActive: true, DisplayName: "Dec",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveUser: %v", err)
	}
	decID := "dec-" + uuid.NewString()
	dec := &AdDeliveryDecision{
		DecisionID: decID, SlotID: "watch_preroll",
		CampaignID: "camp-cloud-001", CreativeID: "creative-preroll-001",
		UserID: ownerID, DecidedAt: time.Now().UTC().Truncate(time.Second),
		Status: "SERVED", Reason: "MATCHED", BidCpmCents: 5500,
	}
	pgCleanup(t, pg,
		[3]string{"ad_delivery_decisions", "decision_id", decID},
		[3]string{"users", "id", ownerID},
	)
	if err := pg.SaveDecision(ctx, dec); err != nil {
		t.Fatalf("SaveDecision: %v", err)
	}
	got, err := pg.LoadDecision(ctx, decID)
	if err != nil {
		t.Fatalf("LoadDecision: %v", err)
	}
	if got == nil {
		t.Fatal("LoadDecision returned nil for saved decision")
	}
	if got.CampaignID != dec.CampaignID || got.CreativeID != dec.CreativeID || got.UserID != dec.UserID {
		t.Fatalf("refs not restored: got %+v", got)
	}
	if got.Status != "SERVED" || got.Reason != "MATCHED" || got.BidCpmCents != 5500 || got.SlotID != "watch_preroll" {
		t.Fatalf("fields not restored: got %+v", got)
	}
	// UUID user lands in the column (FK holds); prefixed refs land in metadata.
	var userCol, campCol sql.NullString
	var meta string
	if err := pg.DB.QueryRowContext(ctx, `SELECT user_id, campaign_id, metadata FROM ad_delivery_decisions WHERE decision_id = $1`, decID).Scan(&userCol, &campCol, &meta); err != nil {
		t.Fatalf("inspect row: %v", err)
	}
	if !userCol.Valid || userCol.String != ownerID {
		t.Fatalf("uuid user should persist in column, got %+v", userCol)
	}
	if campCol.Valid {
		t.Fatalf("prefixed campaign should persist as NULL column, got %q", campCol.String)
	}
	if !strings.Contains(meta, "camp-cloud-001") || !strings.Contains(meta, "creative-preroll-001") {
		t.Fatalf("prefixed refs missing from metadata: %s", meta)
	}
	// Misses report (nil, nil) so event checks fail closed.
	missing, err := pg.LoadDecision(ctx, "dec-does-not-exist")
	if err != nil || missing != nil {
		t.Fatalf("LoadDecision miss = (%v, %v), want (nil, nil)", missing, err)
	}
	recent, err := pg.LoadRecentDecisions(ctx, time.Now().Add(-time.Hour), 10)
	if err != nil {
		t.Fatalf("LoadRecentDecisions: %v", err)
	}
	found := false
	for _, d := range recent {
		if d.DecisionID == decID {
			found = true
		}
	}
	if !found {
		t.Fatal("saved decision missing from recent decisions")
	}
}

func TestSaveDecisionValidation(t *testing.T) {
	pg := &PostgresStore{}
	if err := pg.SaveDecision(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil decision")
	}
	if err := pg.SaveDecision(context.Background(), &AdDeliveryDecision{}); err == nil {
		t.Fatal("expected error for empty decision id")
	}
	if err := pg.SaveDecision(context.Background(), &AdDeliveryDecision{DecisionID: "dec-x"}); err == nil {
		t.Fatal("expected error for empty slot id")
	}
}

// P26: FindDecision serves memory first, restores persisted decisions on a
// miss, and still reports false for unknown IDs (fail-closed).
func TestFindDecisionFallback(t *testing.T) {
	s := NewMemoryStore()
	if _, ok := s.FindDecision("dec-nope"); ok {
		t.Fatal("memory-only miss should report false")
	}
	pg := pgTestStore(t)
	s2 := NewMemoryStoreWithDB(pg)
	decID := "dec-" + uuid.NewString()
	dec := &AdDeliveryDecision{
		DecisionID: decID, SlotID: "watch_preroll",
		CampaignID: "camp-cloud-001", CreativeID: "creative-preroll-001",
		UserID: "user-fb-1", DecidedAt: time.Now().UTC(),
		Status: "SERVED", Reason: "MATCHED", BidCpmCents: 5500,
	}
	pgCleanup(t, pg, [3]string{"ad_delivery_decisions", "decision_id", decID})
	s2.PersistDecision(dec)
	if _, ok := s2.AdDecisions[decID]; ok {
		t.Fatal("precondition: decision should not be in memory")
	}
	got, ok := s2.FindDecision(decID)
	if !ok || got == nil {
		t.Fatal("fallback should restore persisted decision")
	}
	if got.CampaignID != "camp-cloud-001" || got.UserID != "user-fb-1" {
		t.Fatalf("refs not restored: %+v", got)
	}
	if _, ok := s2.AdDecisions[decID]; !ok {
		t.Fatal("fallback should backfill the memory map")
	}
	if _, ok := s2.FindDecision("dec-nope"); ok {
		t.Fatal("unknown id should report false even with DB")
	}
}
