package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	iss    string
}

func NewJWTAuthenticator(secret, aud, iss string) JWTAuthenticator {
	return JWTAuthenticator{secret, aud, iss}
}

func (au *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(au.secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
func (au *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token %v", t.Header["alg"])
		}

		return []byte(au.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(au.aud),
		jwt.WithIssuer(au.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
