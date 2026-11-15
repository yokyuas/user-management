package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() { gin.SetMode(gin.TestMode) }

func signToken(method jwt.SigningMethod, secret string, exp time.Time) string {
	tok := jwt.NewWithClaims(method, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)})
	s, _ := tok.SignedString([]byte(secret))
	return s
}

func TestJWT(t *testing.T) {
	secret := "test-secret"
	validToken := signToken(jwt.SigningMethodHS256, secret, time.Now().Add(time.Hour))
	wrongSecretToken := signToken(jwt.SigningMethodHS256, "wrong-secret", time.Now().Add(time.Hour))
	expiredToken := signToken(jwt.SigningMethodHS256, secret, time.Now().Add(-time.Hour))
	wrongMethodToken := signToken(jwt.SigningMethodHS384, secret, time.Now().Add(time.Hour))

	cases := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
	}{
		{"valid bearer token", "Bearer " + validToken, http.StatusOK, "ok"},
		{"valid lowercase bearer", "bearer " + validToken, http.StatusOK, "ok"},
		{"missing header", "", http.StatusUnauthorized, "missing bearer token"},
		{"non-bearer scheme", "Basic " + validToken, http.StatusUnauthorized, "missing bearer token"},
		{"bearer without space", "Bearer", http.StatusUnauthorized, "missing bearer token"},
		{"malformed token", "Bearer not-a-jwt", http.StatusUnauthorized, "invalid token"},
		{"wrong secret", "Bearer " + wrongSecretToken, http.StatusUnauthorized, "invalid token"},
		{"expired token", "Bearer " + expiredToken, http.StatusUnauthorized, "invalid token"},
		{"unexpected signing method HS384", "Bearer " + wrongMethodToken, http.StatusUnauthorized, "invalid token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(JWT(secret))
			r.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" { req.Header.Set("Authorization", tc.authHeader) }
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus || !strings.Contains(w.Body.String(), tc.wantBody) {
				t.Fatalf("status=%d body=%q, want status=%d body containing %q", w.Code, w.Body.String(), tc.wantStatus, tc.wantBody)
			}
		})
	}
}