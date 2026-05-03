package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type Session struct {
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Service struct {
	adminUsername string
	adminPassword string
	signingKey    []byte
	ttl           time.Duration
	baseURL       string
}

const CookieName = "ten_db_launch_session"

func New(adminUsername, adminPassword, signingSecret, baseURL string, ttl time.Duration) *Service {
	return &Service{
		adminUsername: adminUsername,
		adminPassword: adminPassword,
		signingKey:    []byte(signingSecret),
		ttl:           ttl,
		baseURL:       baseURL,
	}
}

func (s *Service) Login(username, password string) bool {
	return username == s.adminUsername && password == s.adminPassword
}

func (s *Service) CreateCookie() (*http.Cookie, error) {
	session := Session{
		Username:  s.adminUsername,
		ExpiresAt: time.Now().Add(s.ttl).UTC(),
	}
	raw, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	sig := s.sign(payload)
	return &http.Cookie{
		Name:     CookieName,
		Value:    payload + "." + sig,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   strings.HasPrefix(strings.ToLower(s.baseURL), "https://"),
		Expires:  session.ExpiresAt,
	}, nil
}

func (s *Service) ClearCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
}

func (s *Service) Verify(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid session cookie")
	}
	payload, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sig), []byte(s.sign(payload))) {
		return nil, errors.New("invalid session signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, err
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}
	return &session, nil
}

func (s *Service) sign(payload string) string {
	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) BaseURL() string {
	return s.baseURL
}

func Require(service *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := service.Verify(r); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func EnforceSameOrigin(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if origin == "" || !isAllowedOrigin(origin, allowedOrigins) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == strings.TrimRight(allowed, "/") {
			return true
		}
	}
	return false
}
