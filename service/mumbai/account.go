package imone

import (
	"fmt"

	"web-api/model"
)

func (c *Mumbai) CreateWallet(user model.User, currency string) error {
	return c.createImOneUserAndDbWallet(user, currency)
}

// skips duplicate error on insert.
func (c *Mumbai) createImOneUserAndDbWallet(user model.User, currency string) error {
	return fmt.Errorf("not implemented") // @Seng

}

func (c *Mumbai) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	return 0, fmt.Errorf("not implemented") // @Seng
}
