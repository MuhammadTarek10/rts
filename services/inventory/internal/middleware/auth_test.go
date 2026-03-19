package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rts/inventory/internal/middleware"
)

const testSecret = "test-secret-key-for-unit-tests"

func createTestToken(t *testing.T, sub, role string, exp time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"exp":  exp.Unix(),
	})
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestJWTAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		role := middleware.GetRole(r.Context())
		w.Header().Set("X-User-ID", userID)
		w.Header().Set("X-Role", role)
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := middleware.JWTAuth(testSecret)

	t.Run("valid token passes through", func(t *testing.T) {
		token := createTestToken(t, "user-123", "admin", time.Now().Add(time.Hour))
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMiddleware(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("X-User-ID") != "user-123" {
			t.Errorf("expected user-123, got %s", rec.Header().Get("X-User-ID"))
		}
		if rec.Header().Get("X-Role") != "admin" {
			t.Errorf("expected admin, got %s", rec.Header().Get("X-Role"))
		}
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		authMiddleware(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		token := createTestToken(t, "user-123", "admin", time.Now().Add(-time.Hour))
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMiddleware(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("invalid secret returns 401", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  "user-123",
			"role": "admin",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		signed, _ := token.SignedString([]byte("wrong-secret"))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		rec := httptest.NewRecorder()

		authMiddleware(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("cookie token is accepted", func(t *testing.T) {
		token := createTestToken(t, "user-456", "user", time.Now().Add(time.Hour))
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		rec := httptest.NewRecorder()

		authMiddleware(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("X-User-ID") != "user-456" {
			t.Errorf("expected user-456, got %s", rec.Header().Get("X-User-ID"))
		}
	})
}

func TestRequireAdmin(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := middleware.JWTAuth(testSecret)

	t.Run("admin role passes", func(t *testing.T) {
		token := createTestToken(t, "admin-user", "admin", time.Now().Add(time.Hour))
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMiddleware(middleware.RequireAdmin(innerHandler)).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("non-admin role returns 403", func(t *testing.T) {
		token := createTestToken(t, "regular-user", "user", time.Now().Add(time.Hour))
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMiddleware(middleware.RequireAdmin(innerHandler)).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})
}
