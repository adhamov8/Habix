package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"unicode"

	"tracker/internal/service"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Param body object true "Registration data"
// @Success 201 {object} domain.TokenPair
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	if !emailRegex.MatchString(req.Email) {
		jsonError(w, "некорректный формат email", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 2 || len(req.Name) > 50 {
		jsonError(w, "имя должно содержать от 2 до 50 символов", http.StatusBadRequest)
		return
	}
	if !isValidPassword(req.Password) {
		jsonError(w, "пароль должен содержать минимум 8 символов, включая буквы и цифры", http.StatusBadRequest)
		return
	}

	pair, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			jsonError(w, "пользователь с таким email уже существует", http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, pair, http.StatusCreated)
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// Login godoc
// @Summary Login
// @Tags auth
// @Param body body object true "Login credentials"
// @Success 200 {object} domain.TokenPair
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	pair, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			jsonError(w, "неверный email или пароль", http.StatusUnauthorized)
			return
		}
		jsonError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, pair, http.StatusOK)
}

// Refresh godoc
// @Summary Refresh tokens
// @Tags auth
// @Param body body object true "Refresh token"
// @Success 200 {object} domain.TokenPair
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		jsonError(w, "требуется refresh_token", http.StatusBadRequest)
		return
	}

	pair, err := h.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			jsonError(w, "недействительная или устаревшая ссылка", http.StatusUnauthorized)
			return
		}
		jsonError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, pair, http.StatusOK)
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Param body body object true "Refresh token"
// @Success 204
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		jsonError(w, "требуется refresh_token", http.StatusBadRequest)
		return
	}
	_ = h.authSvc.Logout(r.Context(), req.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}
