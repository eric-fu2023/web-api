package imone

import (
	"fmt"
	"web-api/model"
)

func (c *Mumbai) GetGameUrl(user model.User, currency, tayaGameCode, tayaSubGameCode string, _ int64, extra model.Extra) (string, error) {
	return "", fmt.Errorf("not implemented") // @Seng
}
