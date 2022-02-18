package v1

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

func Verify(jwtData string, publicKey *ecdsa.PublicKey) (*jwt.Token, error) {

	claims := ProClaims{}
	parsed, parseErr := jwt.ParseWithClaims(jwtData, &claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if parseErr != nil {
		return nil, fmt.Errorf("JWT parse error: %s", parseErr)
	}

	if !parsed.Valid {
		return nil, fmt.Errorf("jwt invalid")
	}

	return parsed, nil
}

func Issue(privateKey interface{}, name string, email string, products []string, duration time.Duration) (string, error) {
	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)

	id := rand.Intn(10000)

	claims := ProClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%d", id),
			Issuer:    fmt.Sprintf("https://openfaas.com/support/"),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   name,
			Audience:  []string{"https://openfaas.com/support/"},
		},
		Name:     name,
		Email:    email,
		Products: products,
	}

	jwtOutput, err := jwt.NewWithClaims(method, claims).SignedString(privateKey)

	return jwtOutput, err
}
