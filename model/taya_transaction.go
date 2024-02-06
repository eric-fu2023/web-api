package model

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type TayaTransaction struct {
	ploutos.FbTransaction
}

type TayaTransactionClone struct {
	ploutos.FbTransactionClone
}

func (TayaTransaction) TableName() string {
	return "taya_transactions"
}
