package task

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"sync"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

func EncryptMobileAndEmail() {
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
			var toUpdate bool
			if user.MobileHash == "" && user.Mobile != "" {
				if enc, e := util.AesEncrypt([]byte(user.Mobile)); e == nil {
					user.MobileHash = serializer.MobileEmailHash(user.Mobile)
					user.Mobile = enc
					toUpdate = true
				}
			}
			if user.EmailHash == "" && user.Email != "" {
				if enc, e := util.AesEncrypt([]byte(user.Email)); e == nil {
					user.EmailHash = serializer.MobileEmailHash(user.Email)
					user.Email = enc
					toUpdate = true
				}
			}
			if toUpdate {
				model.DB.Updates(&user)
			}
		}(user)
		if i > 99 {
			wg.Wait()
			i = 0
		}
	}
	wg.Wait()
}
