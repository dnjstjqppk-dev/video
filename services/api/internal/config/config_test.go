package config

import (
	"testing"
)

func TestLoadDefaultAndCustomConfig(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	cfg := Load()
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.R2Bucket != "gvp-media-vod" {
		t.Fatalf("expected bucket gvp-media-vod, got %s", cfg.R2Bucket)
	}

	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://custom:pass@db:5432/prod")
	cfgCustom := Load()
	if cfgCustom.Port != "9090" || cfgCustom.DatabaseURL != "postgres://custom:pass@db:5432/prod" {
		t.Fatalf("unexpected custom config: %+v", cfgCustom)
	}
}
