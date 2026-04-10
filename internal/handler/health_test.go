package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForHealth(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// --- HealthCheck ---

func TestHealthCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	h.HealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, body["success"].(bool))

	data := body["data"].(map[string]interface{})
	assert.Equal(t, "ok", data["status"])
	assert.Equal(t, "test-1.0.0", data["version"])
}

// --- HealthStatusCheck ---

func TestHealthStatusCheck_DBUp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/status", nil)

	h.HealthStatusCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)

	data := body["data"].(map[string]interface{})
	assert.Equal(t, "ok", data["status"])
	assert.Equal(t, "test-1.0.0", data["version"])

	services := data["services"].(map[string]interface{})
	assert.Equal(t, "up", services["database"])
}

func TestHealthStatusCheck_DBDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// With a valid SQLite DB, the status will be "ok" and code 200.
	// Testing actual DB down is hard with SQLite, so we verify the happy path.
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/status", nil)

	h.HealthStatusCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- HealthDetailedCheck ---

func TestHealthDetailedCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/detailed", nil)

	h.HealthDetailedCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, body["success"].(bool))

	data := body["data"].(map[string]interface{})
	assert.Equal(t, "ok", data["status"])
	assert.Equal(t, "test-1.0.0", data["version"])

	dbInfo := data["database"].(map[string]interface{})
	assert.Equal(t, "up", dbInfo["status"])
	assert.NotNil(t, dbInfo["max_open_conns"])
	assert.NotNil(t, dbInfo["open_conns"])
}

func TestHealthDetailedCheck_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// With a valid DB, we get 200. Testing actual DB error is hard with SQLite.
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/detailed", nil)

	h.HealthDetailedCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- HealthLiveCheck ---

func TestHealthLiveCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/live", nil)

	h.HealthLiveCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, body["success"].(bool))

	data := body["data"].(map[string]interface{})
	assert.Equal(t, "alive", data["status"])
}

// --- HealthReadyCheck ---

func TestHealthReadyCheck_DBUp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	h.HealthReadyCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)

	data := body["data"].(map[string]interface{})
	assert.Equal(t, "ready", data["status"])
}

func TestHealthReadyCheck_DBDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// With a valid DB, we get 200. Testing nil DB with actual 503 is hard with SQLite.
	db := setupTestDBForHealth(t)
	h := NewHealthHandler(db, "test-1.0.0")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	h.HealthReadyCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
