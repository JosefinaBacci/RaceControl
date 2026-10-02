package transport

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter builds the HTTP routing tree.
//
// Business endpoints arrive with US5 and US6; this version only exposes the
// readiness probe that the container setup and the demo rely on.
func NewRouter(pool *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(requestTimeout))

	router.Get("/healthz", healthHandler(pool))

	return router
}

const requestTimeout = 15 * time.Second

// healthHandler reports readiness by checking that the database answers.
func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, `{"status":"unavailable","database":"unreachable"}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"status":"ok","database":"reachable"}`)
	}
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
