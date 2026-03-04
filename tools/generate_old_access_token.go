package main

import (
	"fmt"
	"os"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	// Run from root directory of the project `go run tools/generate_old_access_token.go`
	configFilePath := "example-config.yaml"
	jwtKey := "-----BEGIN PRIVATE KEY-----\nMIHuAgEAMBAGByqGSM49AgEGBSuBBAAjBIHWMIHTAgEBBEIBE1c1laUlgGPVWVPH\n6jAzB6CcFpPhea12DnfsGZSQ5oOr7hHIfg8ISMCUdtSfFf1VhsO+8eLeJuinp4ro\nr4vMqVChgYkDgYYABABfllyQMGEalHCmMZTKohOHKP3FOhmv4sG7WVZ72Wb1YLqT\noJ7/BhTiRID9OIWB7n78SLQ2xvuCJLlRbHrtqSRoQwF96gLgi7hSBUUP8Sdhhe8y\nrjel1nBKL4NfJWda4hyVEgpiqa9UJIqVkpDi3EciHDYLUMW/pcl78otmhGkncz1+\npg==\n-----END PRIVATE KEY-----"
	err := os.Setenv("SERVER_AUTH_JWT_KEY", jwtKey)
	if err != nil {
		panic(err)
	}

	config, err := config.LoadConfig(configFilePath)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	expiresAt := now.Add(-12 * time.Hour)
	claims := handlers.JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.Nil.String(),
			Issuer:    config.Server.Auth.JwtIssuer,
			IssuedAt:  jwt.NewNumericDate(start),
			NotBefore: jwt.NewNumericDate(start),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	signedToken, err := token.SignedString(config.Server.Auth.JwtEcdsaParsedKey)
	if err != nil {
		panic(err)
	}

	fmt.Println(signedToken)
}
