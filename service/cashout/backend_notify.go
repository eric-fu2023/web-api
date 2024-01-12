package cashout

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

func notifyBackendWithdraw(id string) (err error) {
	_, err = resty.New().R().SetBody(map[string]any{
		"name": "Cash Order To Approve",
		"data": map[string]any{
			"id": id,
		},
	}).Post(fmt.Sprintf("%s/internal/socketMessage", os.Getenv("BACKEND_BASE_URL")))
	return
}
