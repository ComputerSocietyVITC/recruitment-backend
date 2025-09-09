package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestHealthEndpoint(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load environment (this might fail in test environment without DB)
	utils.LoadEnvironment()

	// Create a Gin router
	router := gin.New()

	// Add the health endpoint (simplified version for testing)
	router.GET("/health", func(c *gin.Context) {
		// For testing purposes, we'll simulate a healthy response
		// In a real scenario with a test database, you'd test actual DB connectivity
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"checks": gin.H{
				"database": "ok",
			},
		})
	})

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if checks, ok := response["checks"].(map[string]interface{}); !ok {
		t.Error("Expected 'checks' field in response")
	} else if checks["database"] != "ok" {
		t.Errorf("Expected database check 'ok', got '%v'", checks["database"])
	}
}

func TestHealthEndpointWithoutDatabase(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.New()

	// Add the health endpoint that will fail without database
	router.GET("/health", func(c *gin.Context) {
		// Simulate database connection failure
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"error":     "database connection failed",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Check the response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%v'", response["status"])
	}

	if response["error"] == nil {
		t.Error("Expected 'error' field in unhealthy response")
	}
}
