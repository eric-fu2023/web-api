package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KycDocument struct {
	models.KycDocumentC
}

func GetKycDocumentsWithLock(tx *gorm.DB, kycId int64) ([]KycDocument, error) {
	if tx == nil {
		tx = DB
	}

	var kycDocs []KycDocument
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("kyc_id", kycId).Find(&kycDocs).Error
	if err != nil {
		return nil, err
	}

	return kycDocs, nil
}

func DeleteKycDocuments(tx *gorm.DB, kycId int64) error {
	if tx == nil {
		tx = DB
	}

	return tx.Where("kyc_id", kycId).Delete(&KycDocument{}).Error
}
