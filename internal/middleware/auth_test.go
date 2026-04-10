package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	testUUID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
	claims := JWTClaims{
		UserID: testUUID,
		Email:  "test@example.com",
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+signedToken)

	authMiddleware := Auth(jwtSecret)
	authMiddleware(c)

	assert.False(t, c.IsAborted())
	userID, exists := c.Get("userID")
	assert.True(t, exists)
	assert.Equal(t, testUUID, userID)
}

func TestAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	authMiddleware := Auth(jwtSecret)
	authMiddleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	authMiddleware := Auth(jwtSecret)
	authMiddleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_MalformedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat token")

	authMiddleware := Auth(jwtSecret)
	authMiddleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
