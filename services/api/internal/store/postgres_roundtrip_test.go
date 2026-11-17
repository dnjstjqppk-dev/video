package store

import (
	"context"
	"os"
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
}
