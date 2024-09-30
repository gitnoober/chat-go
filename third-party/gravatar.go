package thirdparty

import (
	"fmt"

	utils "github.com/gitnoober/chat-go/utils"
)

// GetRandomProfilePicture fetches a random gravatar
func GetRandomProfilePicture(email string) string {
	hash := utils.GenerateMD5Hash(email)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=identicon", hash)
}
