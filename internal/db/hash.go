package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func doHash(canonicalJson string) string {
	hash := sha256.Sum256([]byte(canonicalJson))
	return hex.EncodeToString(hash[:]) // [:] converts the array to a slice
}

func HashCreateShortUrlRequest(destinationUrl string) string {
	canonicalJson := fmt.Sprintf(`{"destination_url":"%s"}`, destinationUrl)
	return doHash(canonicalJson)
}

func HashCreateUserRequest(username string, email string) string {
	canonicalJson := fmt.Sprintf(`{"username":"%s","email":"%s"}`, username, email)
	return doHash(canonicalJson)
}
