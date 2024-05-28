package task

import (
	"fmt"
	"web-api/model"
)

func SetRandomAvatar() {
	var users []model.User
	err := model.DB.Model(model.User{}).Where(`avatar = ''`).Order(`id`).Find(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, user := range users {
		model.SetRandomAvatar(&user)
		err = model.DB.Save(&user).Error
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
