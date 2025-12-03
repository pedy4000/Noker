package api

import (
	"net/http"

	"github.com/pedy4000/noker/internal/api/middleware"
	"github.com/pedy4000/noker/pkg/config"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	if cfg.Env == "development" {
		r.Use(chimiddleware.Logger)
	}

	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graph", http.StatusMovedPermanently)
	})
	r.Get("/graph", h.ServeGraph)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))

		// Meetings
		r.Post("/api/meetings", middleware.Validate[CreateMeetingRequest](h.CreateMeeting))
		r.Get("/api/meetings/{id}/status", func(rw http.ResponseWriter, r *http.Request) {
			h.MeetingStatus(rw, r, chi.URLParam(r, "id"))
		})

		// Opportunities
		r.Get("/api/opportunities/recent", h.RecentOpportunities)
		r.Get("/api/opportunities/{id}", func(rw http.ResponseWriter, r *http.Request) {
			h.GetOpportunity(rw, r, chi.URLParam(r, "id"))
		})

		// Themes
		r.Get("/api/themes/{theme}/top-opportunities", func(rw http.ResponseWriter, r *http.Request) {
			h.TopOpportunitiesByTheme(rw, r, chi.URLParam(r, "theme"))
		})
		// Slack command
		r.Get("/api/cmd", h.SlackCommand)
	})

	return r
}
