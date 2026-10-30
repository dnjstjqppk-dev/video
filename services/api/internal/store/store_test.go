package store

import (
	"testing"
)

func TestMemoryStoreInitializationAndSanitizeURL(t *testing.T) {
	ms := NewMemoryStore()
	if ms == nil || len(ms.Users) == 0 || len(ms.Videos) == 0 {
		t.Fatal("expected seeded MemoryStore")
	}

	sanitized := SanitizeDatabaseURL("postgres://admin:secret123@localhost:5432/gvp")
	if sanitized != "postgres://admin:****@localhost:5432/gvp" {
		t.Fatalf("unexpected sanitized URL: %s", sanitized)
	}
}
