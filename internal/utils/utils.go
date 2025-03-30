package utils

import (
	"fmt"
	"time"
)

func GenerateIDForAsset(assetIdentity string) string {
	return fmt.Sprintf("%v-x123", assetIdentity)
}

func HashWithSource(source string) string {
	return fmt.Sprintf("%v@%v", source, time.Now())
}
