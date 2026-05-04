package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"kanalegeleri/go-app/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const adminSessionCookieName = "go_admin_session"

type AuthService struct {
	store         *repository.Store
	sessionSecret string
	isProd        bool
}

func NewAuthService(store *repository.Store, sessionSecret string, isProd bool) *AuthService {
	return &AuthService{
		store:         store,
		sessionSecret: sessionSecret,
		isProd:        isProd,
	}
}

func (s *AuthService) Authenticate(username, password string) bool {
	ctx := context.Background()
	admin, found, err := s.store.GetAdminByUsername(ctx, strings.TrimSpace(username))
	if err != nil || !found {
		return false
	}

	// Bcrypt hash karşılaştırması
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(strings.TrimSpace(password)))
	return err == nil
}

func (s *AuthService) IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(adminSessionCookieName)
	if err != nil {
		return false
	}
	// İmzayı doğrula: kullanıcıadı|imza formatını bekliyoruz
	parts := strings.SplitN(cookie.Value, "|", 2)
	if len(parts) != 2 {
		return false
	}
	username, signature := parts[0], parts[1]
	expectedSignature := s.generateSignature(username)
	return signature == expectedSignature
}

func (s *AuthService) SetSession(w http.ResponseWriter, username string) {
	signature := s.generateSignature(username)
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    fmt.Sprintf("%s|%s", username, signature),
		Path:     "/",
		HttpOnly: true,
		Secure:   s.isProd, // Sadece prod modunda (HTTPS) true olacak
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 8,
	})
}

func (s *AuthService) generateSignature(username string) string {
	h := sha256.New()
	h.Write([]byte(username + s.sessionSecret))
	return fmt.Sprintf("%x", h.Sum(nil))
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
