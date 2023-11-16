package model

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type TayaTransaction struct {
	ploutos.FbTransaction
}

func (TayaTransaction) TableName() string {
	return "taya_transactions"
}
