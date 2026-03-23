package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewEchoApp creates a configured *echo.Echo instance with standard middleware.
func NewEchoApp() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Standard middleware stack
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Health-check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return e
}

// TestNewEchoApp verifies /health returns 200.
func TestNewEchoApp(t *testing.T) {
	e := NewEchoApp()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 from /health, got %d", w.Code)
	}
}
