package v1

import (
	"encoding/base64"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v4"
)

// LoadLicenseToken load a LicenseToken from a string
func LoadLicenseToken(license string, publicKey string) (*LicenseToken, error) {
	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid base64-encoded public-key, error: %s", err)
	}

	key, err := jwt.ParseECPublicKeyFromPEM(decoded)
	if err != nil {
		return nil, fmt.Errorf("tampering detected, please report to contact@openfaas.com, error: %s", err)
	}

	token, err := Verify(license, key)
	if err != nil {
		return nil, fmt.Errorf("invalid license, error: %s", err.Error())
	}

	claims := token.Claims.(*ProClaims)

	return &LicenseToken{
		Name:     claims.Name,
		Email:    claims.Email,
		Products: claims.Products,
		Expires:  claims.ExpiresAt.Time,
	}, nil
}
