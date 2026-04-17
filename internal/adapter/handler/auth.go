package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"go.uber.org/zap"
)

type AuthHandler struct {
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	req, err := parseAuthRequest(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad request")
		return
	}

	token, err := h.uc.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, entity.ErrUserExists) {
			response.WriteError(w, http.StatusConflict, entity.ErrUserExists.Error())
			return
		}
		zap.S().Errorw("failed to register user", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := parseAuthRequest(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad request")
		return
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.uc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) || errors.Is(err, entity.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24,
	})
	w.WriteHeader(http.StatusOK)
}

func parseAuthRequest(r *http.Request) (*authRequest, error) {
	var req authRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&req); err != nil {
		return nil, err
	}

	if req.Login == "" || req.Password == "" {
		return nil, errors.New("empty login or password")
	}
	return &req, nil
}
