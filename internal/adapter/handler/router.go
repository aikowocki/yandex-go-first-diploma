package handler

import (
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.WithLogging())
	r.Use(middleware.WithRecovery)

	r.Route("/api/user", func(r chi.Router) {
		// JSON
		r.Group(func(r chi.Router) {
			r.Use(chimw.AllowContentType("application/json"))
			r.Post("/register", notImplemented)
			r.Post("/login", notImplemented)
			r.Post("/balance/withdraw", notImplemented)
		})
		// text/plain
		r.Group(func(r chi.Router) {
			r.Use(chimw.AllowContentType("text/plain"))
			r.Post("/orders", notImplemented)
		})

		// GET
		r.Get("/orders", notImplemented)
		r.Get("/balance", notImplemented)
		r.Get("/withdrawals", notImplemented)

	})
	return r
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
