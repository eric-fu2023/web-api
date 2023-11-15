package taya

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type TayaTransaction struct {
	ploutos.FbTransactionC
}

func (TayaTransaction) TableName() string {
	return "taya_transactions"
}
