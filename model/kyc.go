package model

import (
	"strings"
	"web-api/conf/consts"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Kyc struct {
	models.KycC
}

func GetKycWithLock(tx *gorm.DB, userId int64) (Kyc, error) {
	if tx == nil {
		tx = DB
	}

	kyc := Kyc{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userId).First(&kyc).Error
	return kyc, err
}

func UpdateKyc(tx *gorm.DB, kyc Kyc) error {
	if tx == nil {
		tx = DB
	}

	// Fields need to be explicitly selected as some fields may be set to their zero/empty values
	selectedFields := []string{
		"first_name", "middle_name", "last_name", "birthday", "document_type", "nationality",
		"current_address", "permanent_address", "employment_type", "income_source", "associated_store", "status",
	}

	res := tx.Select(selectedFields).Updates(&kyc)
	if res.Error != nil {
		util.Log().Error("update err", res.Error)
		return res.Error
	}

	return nil
}

func AcceptKyc(kycId int64) error {
	return DB.Model(Kyc{}).Where(`id`, kycId).Updates(map[string]interface{}{"status": consts.KycStatusCompleted, "remark": ""}).Error
}

func RejectKycWithReason(kycId int64, reason string) error {
	return DB.Model(Kyc{}).Where(`id`, kycId).Updates(map[string]interface{}{"status": consts.KycStatusRejected, "remark": reason}).Error
}

func (k Kyc) NameMatch(name string) bool {
	return name == k.FullName()
}

func (k Kyc) FullName() string {
	list := []string{}
	if k.FirstName != "" {
		list = append(list, k.FirstName)
	}
	if k.MiddleName != "" {
		list = append(list, k.MiddleName)
	}
	if k.LastName != "" {
		list = append(list, k.LastName)
	}
	return strings.Join(list, " ")
}
