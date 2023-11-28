package model

import models "blgit.rfdev.tech/taya/ploutos-object"

func (user *User) GetTagIDList() (list []int64) {
	DB.Table(models.UserTagConn{}.TableName()).Where("user_id", user.ID).Select("user_tag_id").Scan(&list)
	return
}
