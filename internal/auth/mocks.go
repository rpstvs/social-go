package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockAuthenticator struct {
}

const secret = "secret"

var testClaims = jwt.MapClaims{
	"sub": int64(42),
	"exp": time.Now().Add(time.Hour).Unix(),
	"iss": "test-aud",
	"aud": "test-aud",
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{}
}

func (a *MockAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *MockAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

}
