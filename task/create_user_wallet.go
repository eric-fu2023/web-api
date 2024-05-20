package task

import (
	"fmt"
	"strconv"
	"sync"

	"web-api/model"
	"web-api/service/imone"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

func CreateUserWallet(gameVendorIds []int64, gameIntegrationId int64) {
	var users []ploutos.User
	err := model.DB.Model(ploutos.User{}).Order(`id`).Find(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}

	var cgi []ploutos.CurrencyGameIntegration
	err = model.DB.Model(ploutos.CurrencyGameIntegration{}).Where(`game_integration_id`, gameIntegrationId).Find(&cgi).Error
	if err != nil {
		fmt.Println(err)
		return
	}

	currMap := make(map[int64]string)
	for _, c := range cgi {
		if c.CurrencyId != 0 {
			currMap[c.CurrencyId] = c.Value
		}
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
					fmt.Println("user's wallet exists, skipping")
					continue
				}
				currency, exists := currMap[user.CurrencyId]
				if !exists {
					fmt.Println("user's currency id invalid")
					return
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

// assumes if game_vendor_user not created => yet to register with imone
func CreateImOneUsersForExistingTayaUsers() {
	currency := "INR"
	var userIds []int64
	tx := model.DB.Raw(fmt.Sprintf("select user_id from user_sums where user_id  not in (select user_id from game_vendor_users gvu, game_vendor gv, game_integrations gi where gvu.game_vendor_id = gv.id and gv.game_integration_id = gi.id and gi.name = 'IMONE');")).Find(&userIds)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}
	fmt.Printf("userIds: %v", len(userIds))

	service := &imone.ImOne{}

	var wg sync.WaitGroup
	for _, userId := range userIds {
		wg.Add(1)
		go func(userId int64) {
			defer wg.Done()
			err := service.CreateWallet(model.User{
				User: ploutos.User{
					BASE: ploutos.BASE{
						ID: userId,
					},
				},
			}, currency)

			if err != nil {
				fmt.Println("err creating waller. user id " + strconv.Itoa(int(userId)) + " err: " + err.Error())
			} else {
				fmt.Println("ok creating waller. user id " + strconv.Itoa(int(userId)) + " ")
			}
		}(userId)
	}

	wg.Wait()

}
