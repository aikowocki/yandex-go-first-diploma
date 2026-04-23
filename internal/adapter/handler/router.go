package handler

import (
	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler/middleware"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(auth *AuthHandler, order *OrderHandler, balance *BalanceHandler, health *HealthHandler, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(chimw.RequestID)
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
				r.Post("/balance/withdraw", balance.Withdraw)
			})
			// text/plain
			r.Group(func(r chi.Router) {
				r.Use(chimw.AllowContentType("text/plain"))
				r.Post("/orders", order.SubmitOrder)
			})

			// GET
			r.Get("/orders", order.GetUserOrders)
			r.Get("/balance", balance.GetBalance)
			r.Get("/withdrawals", balance.GetWithdrawals)
		})

	})
	r.Get("/ping", health.Ping)
	return r
}
