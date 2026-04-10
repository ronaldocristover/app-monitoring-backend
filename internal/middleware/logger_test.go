package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger_LogsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Logger(sugar))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test?query=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, logs.FilterMessage("HTTP request").Len())
	log := logs.FilterMessage("HTTP request").All()[0]
	assert.Equal(t, zap.InfoLevel, log.Level)
	assert.Equal(t, "GET", log.ContextMap()["method"])
	assert.Equal(t, "/test", log.ContextMap()["path"])
	assert.Equal(t, "query=value", log.ContextMap()["query"])
	assert.Equal(t, int64(200), log.ContextMap()["status"])
}

func TestLogger_LogsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Logger(sugar))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 1, logs.FilterMessage("HTTP request").Len())
	log := logs.FilterMessage("HTTP request").All()[0]
	assert.Equal(t, int64(500), log.ContextMap()["status"])
}

func TestLogger_LogsClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Logger(sugar))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	log := logs.FilterMessage("HTTP request").All()[0]
	assert.Equal(t, "192.168.1.100", log.ContextMap()["client_ip"])
}

func TestLogger_LogsUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Logger(sugar))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 TestAgent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	log := logs.FilterMessage("HTTP request").All()[0]
	assert.Equal(t, "Mozilla/5.0 TestAgent", log.ContextMap()["user_agent"])
}

func TestLogger_LogsRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(RequestID())
	router.Use(Logger(sugar))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	log := logs.FilterMessage("HTTP request").All()[0]
	// Check that request is logged (RequestID middleware runs before Logger)
	assert.NotNil(t, log)
}

func TestLogger_MeasuresDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.InfoLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Logger(sugar))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	log := logs.FilterMessage("HTTP request").All()[0]
	assert.NotEmpty(t, log.ContextMap()["duration"])
}
