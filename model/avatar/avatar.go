package avatar

import (
	"math/rand"
	"os"
)

var defaultUrlList = urlList1

func GetAvatarUrls() []string {
	lName := os.Getenv("AVATAR_URL_LIST_NAME") // todo lift up to Init
	switch lName {
	case "2":
		return urlList2
	default:
		return defaultUrlList
	}
}

func GetRandomAvatarUrl() string {
	urls := GetAvatarUrls()
	n := rand.Intn(len(urls))
	return urls[n]
}
