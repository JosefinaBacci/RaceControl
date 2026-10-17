package configtest

import (
	"github.com/racecontrol/backend/internal/config"
	"testing"
)

func TestProductionRequiresAnExplicitHTTPSConfiguration(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://unused")
	t.Setenv("HTTP_ADDR", ":8443")
	t.Setenv("SEED_DEMO_PASSWORD", "")
	t.Setenv("TLS_CERT_FILE", "")
	t.Setenv("TLS_KEY_FILE", "")
	t.Setenv("TRUST_PROXY_HEADERS", "false")
	if _, err := config.Load(); err == nil {
		t.Fatal("production without HTTPS accepted")
	}
	t.Setenv("TLS_CERT_FILE", "certificate.pem")
	if _, err := config.Load(); err == nil {
		t.Fatal("certificate without key accepted")
	}
	t.Setenv("TLS_KEY_FILE", "private.pem")
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TLS_CERT_FILE", "")
	t.Setenv("TLS_KEY_FILE", "")
	t.Setenv("TRUST_PROXY_HEADERS", "true")
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
}
