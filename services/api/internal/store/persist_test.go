package store

import (
	"testing"
	"time"
)

func TestPersistWithoutDBIsNoop(t *testing.T) {
	s := NewMemoryStore()
	if s.DB != nil {
		t.Fatal("expected nil DB in unit-test store")
	}
	s.PersistVideo(&Video{ID: "v1"})
	s.PersistUser(&User{ID: "u1"})
	s.PersistTransaction(&PaymentTransaction{ID: "t1", CreatedAt: time.Now()})
	s.PersistVideo(nil)
	s.PersistUser(nil)
	s.PersistTransaction(nil)
}
