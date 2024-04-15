package task

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"sync"
	"web-api/model"
)

func CreateUserWallet(gameVendorIds []int64, currency string) {
	var users []ploutos.User
	err := model.DB.Model(ploutos.User{}).Order(`id`).Find(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	var wg sync.WaitGroup
	var i int
	for _, user := range users {
		i++
		wg.Add(1)
		go func(user ploutos.User) {
			defer wg.Done()
			for _, gameVendorId := range gameVendorIds {
				var gameVendorUser ploutos.GameVendorUser
				rows := model.DB.Where(`user_id`, user.ID).Where(`game_vendor_id`, gameVendorId).First(&gameVendorUser).RowsAffected
				if rows > 0 {
					continue
				}
				gameVendorUser = ploutos.GameVendorUser{
					GameVendorId:     gameVendorId,
					UserId:           user.ID,
					ExternalCurrency: currency,
				}
				err = model.DB.Create(&gameVendorUser).Error
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}(user)
		if i > 99 {
			wg.Wait()
			i = 0
		}
	}
	wg.Wait()
}
