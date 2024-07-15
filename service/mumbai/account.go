package imone

import (
	"fmt"
	"web-api/util"

	"web-api/model"
)

func (c *Mumbai) CreateWallet(user model.User, currency string) error {
	return c.createMumbaiUserAndDbWallet(user, currency)
}

// skips duplicate error on insert.
func (c *Mumbai) createMumbaiUserAndDbWallet(user model.User, currency string) error {
	client, _ := util.MumbaiFactory()
	_ = client
	// call game service :client.LoginUser()
	return fmt.Errorf("not implemented") // @Seng

}

func (c *Mumbai) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	return 0, fmt.Errorf("not implemented") // @Seng
}
