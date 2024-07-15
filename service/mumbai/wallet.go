package imone

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"web-api/model"
	"web-api/util"

	"gorm.io/gorm"
)

// TransferFrom
func (c *Mumbai) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) error {
	client, _ := util.MumbaiFactory()
	_ = client
	return fmt.Errorf("not implemented")
}

func (c *Mumbai) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, _currency, gameCode string, gameVendorId int64, extra model.Extra) (_transferredBalance int64, _err error) {
	return 0, fmt.Errorf("not implemented")
}
