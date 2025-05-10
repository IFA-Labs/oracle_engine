package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func HashWithSource(source string) string {
	return fmt.Sprintf("%v@%v", source, time.Now())
}

// Normalize and hash asset symbol with a namespace
func GenerateIDForAsset(assetIdentity string) string {
	normalized := strings.ToUpper(strings.TrimSpace(assetIdentity))
	seed := "oracle.asset:" + normalized
	hash := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(hash[:])
}
