package transporttest

import (
	"github.com/racecontrol/backend/internal/transport"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductionRejectsPlainHTTPAndUntrustedForwardedHeaders(t *testing.T) {
	router := transport.NewRouter(transport.Dependencies{Cookie: transport.CookieConfig{Secure: true}})
	for _, forwarded := range []string{"", "https"} {
		request := httptest.NewRequest(http.MethodGet, "http://example.test/missing", nil)
		request.Header.Set("X-Forwarded-Proto", forwarded)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUpgradeRequired {
			t.Fatalf("plain HTTP: %d", response.Code)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "https://example.test/missing", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("HTTPS: %d", response.Code)
	}
}

func TestTrustedHTTPSProxyIsAccepted(t *testing.T) {
	router := transport.NewRouter(transport.Dependencies{Cookie: transport.CookieConfig{Secure: true}, TrustProxyHeaders: true})
	request := httptest.NewRequest(http.MethodGet, "http://internal.test/missing", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("trusted HTTPS: %d", response.Code)
	}
}
