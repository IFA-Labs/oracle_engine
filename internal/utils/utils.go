package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"oracle_engine/internal/logging"
	"strings"
	"time"

	"go.uber.org/zap"
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

func Float64ToBigInt(f float64) *big.Int {
	b := new(big.Float).SetFloat64(f)

	// Truncate any fractional part (shouldn’t exist if you’re storing full value)
	result := new(big.Int)
	b.Int(result)

	return result
}

func HexToBytes32(hexStr string) [32]byte {
	var key [32]byte

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		logging.Logger.Warn("invalid hex string: %v", zap.Any("error", err))
	}
	if len(bytes) != 32 {
		logging.Logger.Warn("expected 32 bytes, got %d", zap.Any("length", len(bytes)))
	}

	copy(key[:], bytes)
	return key
}
