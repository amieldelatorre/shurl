package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HashCreateShortUrlRequest(destinationUrl string) string {
	canonicalJson := fmt.Sprintf(`{"destination_url":"%s"}`, destinationUrl)
	hash := sha256.Sum256([]byte(canonicalJson))
	return hex.EncodeToString(hash[:]) // [:] converts the array to a slice
}
