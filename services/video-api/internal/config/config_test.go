package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaultsAndValidOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("MAX_UPLOAD_BYTES", "123")
	t.Setenv("UPLOAD_URL_TTL", "30s")
	t.Setenv("CORS_ORIGIN", "https://one.example,https://two.example")

	cfg := Load()
	if cfg.HTTPAddr != ":9090" || cfg.MaxUploadBytes != 123 || cfg.PresignTTL != 30*time.Second {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	origins := AllowedOrigins(cfg.CORSOrigin)
	if len(origins) != 2 || origins[1] != "https://two.example" {
		t.Fatalf("unexpected origins: %#v", origins)
	}
}

func TestLoadRejectsInvalidNumericAndDurationValues(t *testing.T) {
	t.Setenv("MAX_UPLOAD_BYTES", "-1")
	t.Setenv("UPLOAD_URL_TTL", "not-a-duration")
	cfg := Load()
	if cfg.MaxUploadBytes != 524288000 {
		t.Fatalf("got max upload %d", cfg.MaxUploadBytes)
	}
	if cfg.PresignTTL != 15*time.Minute {
		t.Fatalf("got ttl %s", cfg.PresignTTL)
	}
}
