package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type jwtAuthenticator struct {
	secret string
	aud    string
	issuer string
}

func NewJwtAuthenticator(secret, aud, issuer string) *jwtAuthenticator {
	return &jwtAuthenticator{
		secret: secret,
		aud:    aud,
		issuer: issuer,
	}
}

func (a *jwtAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *jwtAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}
