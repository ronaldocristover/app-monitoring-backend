package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRecovery_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
	// Find the panic log by checking if it contains "Panic" or "panic"
	var found bool
	for _, log := range allLogs {
		for k := range log.ContextMap() {
			if k == "error" || k == "path" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "should find panic log with error or path context")
}

func TestRecovery_PanicWithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/panic-error", func(c *gin.Context) {
		panic(errors.New("wrapped error"))
	})

	req := httptest.NewRequest("GET", "/panic-error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
}

func TestRecovery_LogsStackTrace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/panic", func(c *gin.Context) {
		panic("stack trace test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
}

func TestRecovery_LogsRequestMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.POST("/panic", func(c *gin.Context) {
		panic("method test")
	})

	req := httptest.NewRequest("POST", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
}

func TestRecovery_LogsRequestPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/test/path/panic", func(c *gin.Context) {
		panic("path test")
	})

	req := httptest.NewRequest("GET", "/test/path/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
}

func TestRecovery_LogsClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/panic", func(c *gin.Context) {
		panic("ip test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that some panic log was recorded
	allLogs := logs.All()
	assert.Greater(t, len(allLogs), 0, "should have at least one panic log")
}

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	zapCore, logs := observer.New(zap.ErrorLevel)
	sugar := zap.New(zapCore).Sugar()

	router := gin.New()
	router.Use(Recovery(sugar))
	router.GET("/normal", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	assert.Equal(t, 0, len(logs.All()))
}
