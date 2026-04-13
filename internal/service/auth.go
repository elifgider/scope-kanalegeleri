package service

import (
	"net/http"
	"strings"
)

const adminSessionCookieName = "go_admin_session"

type AuthService struct {
	username string
	password string
}

func NewAuthService(username, password string) *AuthService {
	return &AuthService{
		username: strings.TrimSpace(username),
		password: strings.TrimSpace(password),
	}
}

func (s *AuthService) Authenticate(username, password string) bool {
	return strings.TrimSpace(username) == s.username && strings.TrimSpace(password) == s.password
}

func (s *AuthService) IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(adminSessionCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == s.password
}

func (s *AuthService) SetSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    s.password,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 8,
	})
}

func (s *AuthService) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
