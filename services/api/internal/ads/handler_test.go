package ads

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestAdDecisionAndCampaignRoutes(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	// 1. Test Ad Decision endpoint
	body, _ := json.Marshal(map[string]interface{}{
		"slot_id":  "watch_preroll",
		"user_id":  "33333333-3333-3333-3333-333333333333",
		"category": "Technology",
	})
	req := httptest.NewRequest(http.MethodPost, "/decision", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /decision, got %d", rec.Code)
	}

	// 2. Campaign management requires an authenticated advertiser role (P0-8).
	req2 := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 anonymous from /campaigns, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
	req3.Header.Set("Authorization", "Bearer advertiser-token")
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /campaigns for advertiser, got %d", rec3.Code)
	}
}

func TestCreateCampaign_RejectsNegativeBudgets(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))
	tok := "advertiser" + "-token"
	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/campaigns", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	if rec := post(`{"name":""}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty name, got %d", rec.Code)
	}
	if rec := post(`{"name":"x","dailyBudget":-5}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative budget, got %d", rec.Code)
	}
	if rec := post(`{"name":"x","totalBudget":100,"dailyBudget":10,"cpmBidCents":5,"cpcBidCents":1}`); rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid campaign, got %d: %s", rec.Code, rec.Body.String())
	}
}
