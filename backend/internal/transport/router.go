package transport

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/domain"
)

const requestTimeout = 15 * time.Second

type SessionService interface {
	auth.SessionResolver
	SessionRevoker
}

type Dependencies struct {
	Pool              *pgxpool.Pool
	Authenticator     Authenticator
	Sessions          SessionService
	Users             UserManager
	Teams             TeamCatalog
	Calendar          CalendarManager
	Drivers           DriverManager
	Scores            ScoreBoard
	ScoreNotices      ScoreNoticeManager
	Messenger         Messenger
	Cookie            CookieConfig
	AllowedOrigins    []string
	TrustProxyHeaders bool
}

func NewRouter(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	if deps.TrustProxyHeaders {
		router.Use(middleware.RealIP)
	}
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(requestTimeout))
	router.Use(securityHeaders)
	router.Use(corsFor(deps.AllowedOrigins))

	sessionGuard := auth.NewMiddleware(deps.Sessions, writeError)
	authRoutes := &authHandlers{authenticator: deps.Authenticator, sessions: deps.Sessions, cookie: deps.Cookie}

	router.Get("/healthz", healthHandler(deps.Pool))

	router.Get("/teams", listTeamsHandler(deps.Teams))

	calendarRoutes := &calendarHandlers{calendar: deps.Calendar}
	router.Get("/circuits", calendarRoutes.circuits)
	router.Get("/seasons", calendarRoutes.seasons)
	router.Get("/calendar.ics", calendarRoutes.downloadICS)
	router.Get("/calendar.pdf", calendarRoutes.downloadPDF)

	scoreRoutes := &scoreHandlers{scores: deps.Scores}

	router.Route("/events", func(r chi.Router) {
		r.Get("/", calendarRoutes.list)
		r.Get("/{id}", calendarRoutes.get)
		r.Group(func(r chi.Router) {
			r.Use(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleFIAAdmin))
			r.Post("/", calendarRoutes.create)
			r.Patch("/{id}", calendarRoutes.update)
			r.Delete("/{id}", calendarRoutes.cancel)
			r.Get("/{id}/scores", scoreRoutes.grid)
			r.Put("/{id}/scores", scoreRoutes.replace)
		})
	})

	userRoutes := &userHandlers{users: deps.Users}
	router.Route("/users", func(r chi.Router) {
		r.Use(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleFIAAdmin))
		r.Get("/", userRoutes.list)
		r.Post("/", userRoutes.create)
		r.Get("/{id}", userRoutes.get)
		r.Patch("/{id}", userRoutes.update)
		r.Post("/{id}/deactivate", userRoutes.deactivate)
		r.Post("/{id}/reactivate", userRoutes.reactivate)
	})
	router.With(sessionGuard.RequireSession).Post("/account/password", userRoutes.changeOwnPassword)

	driverRoutes := &driverHandlers{drivers: deps.Drivers}
	router.Get("/drivers", driverRoutes.directory)
	router.Get("/drivers/{id}", driverRoutes.show)
	router.Route("/team/drivers", func(r chi.Router) {
		r.Use(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleTeamAdmin))
		r.Get("/", driverRoutes.list)
		r.Post("/", driverRoutes.create)
		r.Get("/{id}", driverRoutes.get)
		r.Patch("/{id}", driverRoutes.update)
		r.Post("/{id}/release", driverRoutes.release)
	})

	router.Get("/results", scoreRoutes.results)
	router.Get("/results/{id}", scoreRoutes.raceResults)
	router.Get("/standings", scoreRoutes.standings)

	noticeRoutes := &scoreNoticeHandlers{notices: deps.ScoreNotices}
	router.Route("/team/score-notices", func(r chi.Router) {
		r.Use(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleTeamAdmin))
		r.Get("/", noticeRoutes.teamNotices)
		r.Post("/{id}/acknowledgement", noticeRoutes.acknowledge)
	})
	router.With(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleFIAAdmin)).
		Get("/score-changes", noticeRoutes.changes)

	messagingRoutes := &messagingHandlers{messenger: deps.Messenger}
	router.Route("/conversations", func(r chi.Router) {
		r.Use(sessionGuard.RequireSession, sessionGuard.RequireRole(domain.RoleFIAAdmin, domain.RoleTeamAdmin))
		r.Get("/", messagingRoutes.list)
		r.Post("/", messagingRoutes.start)
		r.Get("/{id}", messagingRoutes.open)
		r.Post("/{id}/messages", messagingRoutes.reply)
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", authRoutes.login)
		r.Post("/logout", authRoutes.logout)
		r.With(sessionGuard.RequireSession).Get("/me", authRoutes.me)
	})

	return router
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "database": "unreachable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "reachable"})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		headers.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func corsFor(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowedHeaders:   []string{"Content-Type", "Authorization", clientPlatformHeader},
		AllowCredentials: true,
		MaxAge:           int((10 * time.Minute).Seconds()),
	})
}
