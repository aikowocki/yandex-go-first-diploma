package handler

import (
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler/middleware"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(auth *AuthHandler, order *OrderHandler, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.WithLogging())
	r.Use(middleware.WithRecovery)

	r.Route("/api/user", func(r chi.Router) {
		// PUBLIC
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)

		// PROTECTED
		r.Group(func(r chi.Router) {
			r.Use(middleware.WithAuth(jwtManager))
			// JSON
			r.Group(func(r chi.Router) {
				r.Use(chimw.AllowContentType("application/json"))
				r.Post("/balance/withdraw", notImplemented)
			})
			// text/plain
			r.Group(func(r chi.Router) {
				r.Use(chimw.AllowContentType("text/plain"))
				r.Post("/orders", order.SubmitOrder)
			})

			// GET
			r.Get("/orders", order.GetUserOrders)
			r.Get("/balance", notImplemented)
			r.Get("/withdrawals", notImplemented)
		})

	})
	return r
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
