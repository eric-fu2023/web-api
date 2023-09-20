package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserProfile struct {
	ploutos.UserProfile
	User *User
}
