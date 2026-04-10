package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
)

func TestRequestID_GeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("requestID")
		assert.True(t, exists, "requestID should exist")
		assert.NotEmpty(t, requestID, "requestID should not be empty")
		c.String(http.StatusOK, requestID.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	existingID := "custom-request-id-12345"

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("requestID")
		assert.Equal(t, existingID, requestID, "should use existing header value")
		c.String(http.StatusOK, requestID.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, existingID, w.Body.String())
}

func TestRequestID_HeaderIsSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("requestID")
		c.Header("X-Request-ID", requestID.(string))
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestID_UUIDFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("requestID")
		c.String(http.StatusOK, requestID.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	id := w.Body.String()

	uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, err := regexp.MatchString(uuidRegex, id)
	require.NoError(t, err)
	assert.True(t, matched, "request ID should be UUID format")
}

func TestRequestID_DifferentForEachRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())

	var ids []string
	paths := []string{"/test1", "/test2", "/test3", "/test4", "/test5"}
	for _, path := range paths {
		router.GET(path, func(c *gin.Context) {
			requestID, _ := c.Get("requestID")
			ids = append(ids, requestID.(string))
			c.String(http.StatusOK, requestID.(string))
		})

		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	assert.Len(t, ids, 5)
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			assert.NotEqual(t, ids[i], ids[j], "IDs should be unique")
		}
	}
}

func TestRequestID_PassesThroughMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var middlewareID string

	router := gin.New()
	router.Use(RequestID())
	router.Use(func(c *gin.Context) {
		requestID, exists := c.Get("requestID")
		assert.True(t, exists, "request_id should exist in middleware")
		middlewareID = requestID.(string)
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		handlerID, _ := c.Get("requestID")
		assert.Equal(t, middlewareID, handlerID, "ID should be same through middleware chain")
		c.String(http.StatusOK, handlerID.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, middlewareID, w.Body.String())
}
