package thirdparty

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// Generate MD5 hash of the email string
func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GetRandomProfilePicture fetches a random gravatar
func GetRandomProfilePicture(email string) string {
	hash := md5Hash(email)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=identicon", hash)
}
