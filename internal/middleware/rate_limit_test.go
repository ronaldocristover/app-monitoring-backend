package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_AllowsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}
}

func TestRateLimit_PerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	makeRequest := func(ip string) int {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		req.RemoteAddr = ip + ":12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	for i := 0; i < 5; i++ {
		assert.Equal(t, http.StatusOK, makeRequest("192.168.1.1"))
	}

	assert.Equal(t, http.StatusOK, makeRequest("192.168.1.2"), "different IP should not be rate limited")
	assert.Equal(t, http.StatusOK, makeRequest("192.168.1.3"), "another different IP should not be rate limited")
}

func TestRateLimit_Concurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	var successCount atomic.Int32

	for i := 0; i < 50; i++ {
		go func(ipNum int) {
			req := httptest.NewRequest("GET", "http://example.com/test", nil)
			req.RemoteAddr = "127.0.0." + string(rune('0'+ipNum)) + ":12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				successCount.Add(1)
			}
		}(i % 10)
	}

	time.Sleep(100 * time.Millisecond)
	assert.Greater(t, successCount.Load(), int32(0), "at least some requests should succeed")
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, time.Minute, rl.window)
	rl.Stop()
}

func TestRateLimiter_GetVisitor(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	defer rl.Stop()

	v1 := rl.getVisitor("127.0.0.1")
	assert.NotNil(t, v1)
	assert.Equal(t, 1, v1.count)

	v2 := rl.getVisitor("127.0.0.1")
	assert.Equal(t, v1, v2)
	assert.Equal(t, 2, v2.count)

	v3 := rl.getVisitor("127.0.0.2")
	assert.NotNil(t, v3)
	assert.Equal(t, 1, v3.count)
	assert.NotEqual(t, v2, v3)
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(5, 100*time.Millisecond)
	defer rl.Stop()

	rl.getVisitor("127.0.0.1")
	rl.getVisitor("127.0.0.2")
	rl.getVisitor("127.0.0.3")

	assert.Equal(t, 3, len(rl.visitors))

	time.Sleep(200 * time.Millisecond)
	rl.cleanup()

	assert.Equal(t, 0, len(rl.visitors), "visitors should be cleaned up")
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	assert.NotNil(t, rl.stopCh)

	rl.Stop()
}
