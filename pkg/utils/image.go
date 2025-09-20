package utils

import (
	"strings"

	"github.com/google/uuid"
)

// ExtractImageIDFromKey extracts image ID from S3 key format: images/{userID}/{imageID}.{ext}
func ExtractImageIDFromKey(s3Key string) string {
	parts := strings.Split(s3Key, "/")
	if len(parts) >= 3 {
		filename := parts[len(parts)-1]
		// Remove extension to get image ID
		if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
			return filename[:dotIndex]
		}
		return filename
	}
	return uuid.New().String() // Fallback
}