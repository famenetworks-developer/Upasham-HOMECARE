package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"upasham-backend/internal/auth"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.NewRateLimiter(2, time.Minute).Middleware("test"))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for i, want := range []int{204, 204, 429} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(recorder, request)
		if recorder.Code != want { t.Fatalf("request %d: got %d, want %d", i+1, recorder.Code, want) }
	}
}

func TestSecurityHeadersAndCorrelationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.SecurityHeaders())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, header := range []string{"X-Correlation-ID", "X-Content-Type-Options", "Content-Security-Policy", "Referrer-Policy"} {
		if recorder.Header().Get(header) == "" { t.Fatalf("expected %s header", header) }
	}
}
