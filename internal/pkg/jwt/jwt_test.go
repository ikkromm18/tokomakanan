package jwt_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
)

const (
	testSecret = "test-secret-key-that-is-at-least-32-bytes-long!"
)

func TestGenerateToken_And_ValidateToken_Success(t *testing.T) {
	userID := uint64(42)
	role := "admin"
	expiryHours := 24

	token, err := jwtpkg.GenerateToken(userID, role, testSecret, expiryHours)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := jwtpkg.ValidateToken(token, testSecret)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidateToken_Expired(t *testing.T) {
	userID := uint64(1)
	role := "superadmin"
	expiryHours := -1 // expired 1 hour ago

	token, err := jwtpkg.GenerateToken(userID, role, testSecret, expiryHours)
	require.NoError(t, err)

	claims, err := jwtpkg.ValidateToken(token, testSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	userID := uint64(10)
	role := "owner"

	token, err := jwtpkg.GenerateToken(userID, role, testSecret, 1)
	require.NoError(t, err)

	claims, err := jwtpkg.ValidateToken(token, "different-secret-key-32-chars-long-123456")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_Malformed(t *testing.T) {
	claims, err := jwtpkg.ValidateToken("not.a.valid.jwt.token", testSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_NoneAlgorithmRejected(t *testing.T) {
	// Attacker crafting alg: none
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwtpkg.JWTClaims{
		UserID: 1,
		Role:   "superadmin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	claims, err := jwtpkg.ValidateToken(tokenString, testSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestPasswordHash_And_Compare(t *testing.T) {
	password := "SecretPassword123!"

	hash, err := jwtpkg.HashPassword(password, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Valid password check
	assert.True(t, jwtpkg.CheckPassword(password, hash))
	assert.NoError(t, jwtpkg.ComparePassword(hash, password))

	// Invalid password check
	assert.False(t, jwtpkg.CheckPassword("WrongPassword123!", hash))
	assert.Error(t, jwtpkg.ComparePassword(hash, "WrongPassword123!"))
}
