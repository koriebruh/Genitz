package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// SetupGinRouter creates a production-ready *gin.Engine.
// Call gin.SetMode(gin.TestMode) before this in tests.
func SetupGinRouter() *gin.Engine {
	router := gin.New()

	// Standard middleware
	router.Use(gin.Recovery())

	// Health-check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

// TestSetupGinRouter verifies the router starts up and /health returns 200.
func TestSetupGinRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupGinRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 from /health, got %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}
}
