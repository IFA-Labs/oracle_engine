package utils

import "fmt"

func GenerateIDForAsset(assetIdentity string) string {
	return fmt.Sprintf("%v-x123", assetIdentity)
}
