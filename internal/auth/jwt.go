package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	issuer string
}

func NewJWTAuthenticator(secret, aud, issuer string) *JWTAuthenticator {
	return &JWTAuthenticator{secret, aud, issuer}
}

func (j *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTAuthenticator) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(j.aud),
		jwt.WithIssuer(j.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
