package auth

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	appauth "github.com/emremy/socialqueue/internal/app/auth"
	response "github.com/emremy/socialqueue/internal/transport/httpserver/response"
)

type Handler struct {
	authService *appauth.Service
}

func NewHandler(authService *appauth.Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.authService.Register(
		r.Context(),
		req.Email,
		req.Password,
		req.FullName,
		getUserAgent(r),
		getIPAddress(r),
	)
	if err != nil {
		if errors.Is(err, appauth.ErrEmailAlreadyExists) {
			response.WriteJSON(w, http.StatusConflict, map[string]string{
				"error": "email already exists",
			})
			return
		}

		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	response.WriteJSON(w, http.StatusCreated, NewAuthResponse(*result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.authService.Login(
		r.Context(),
		req.Email,
		req.Password,
		getUserAgent(r),
		getIPAddress(r),
	)
	if err != nil {
		if errors.Is(err, appauth.ErrInvalidCredentials) {
			response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid credentials",
			})
			return
		}

		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, NewAuthResponse(*result))
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.authService.Refresh(
		r.Context(),
		req.RefreshToken,
		getUserAgent(r),
		getIPAddress(r),
	)
	if err != nil {
		if errors.Is(err, appauth.ErrInvalidCredentials) {
			response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid refresh token",
			})
			return
		}

		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, NewAuthResponse(*result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.authService.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, appauth.ErrInvalidCredentials) || errors.Is(err, appauth.ErrInactiveUser) {
			response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, NewUserResponse(*user))
}

func getUserAgent(r *http.Request) *string {
	userAgent := strings.TrimSpace(r.UserAgent())

	if userAgent == "" {
		return nil
	}

	return &userAgent
}

func getIPAddress(r *http.Request) *string {
	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		}
	}

	ip = strings.TrimSpace(strings.Split(ip, ",")[0])

	if ip == "" {
		return nil
	}

	return &ip
}
