package service

import (
	"errors"
	"strconv"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

func VerifyKyc(c *gin.Context, userID int64) (kyc model.Kyc, err error) {
	value, err := GetCachedConfig(c, consts.ConfigKeyTopupKycCheck)
	if err != nil {
		return
	}
	required, err := strconv.ParseBool(value)
	if err != nil {
		return
	}
	if required {
		kyc, err = model.GetKycWithLock(model.DB, userID)
		if err != nil {
			return
		}
		if kyc.Status != consts.KycStatusCompleted {
			err = errors.New("kyc not completed")
			return
		}
	}
	return
}

func VerifyKycWithName(c *gin.Context, userID int64, name string) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	var kyc model.Kyc
	kyc, err = VerifyKyc(c, userID)
	if err != nil {
		r = serializer.Err(c, nil, serializer.CodeGeneralError, i18n.T("kyc_get_failed"), err)
		return
	}
	if !kyc.NameMatch(name) {
		r = serializer.Err(c, nil, serializer.CodeGeneralError, i18n.T("kyc_get_failed"), err)
		return 
	}
	return
}
