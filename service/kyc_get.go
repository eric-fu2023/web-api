package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type GetKycService struct{}

func (service *GetKycService) GetKyc(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)

	kyc := model.Kyc{
		Kyc: ploutos.Kyc{
			UserId: user.ID,
		},
	}
	res := model.DB.Where(&kyc).First(&kyc)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return serializer.Response{}
		} else {
			util.GetLoggerEntry(c).Errorf("Find kyc error: %s", res.Error.Error())
			return serializer.DBErr(c, service, i18n.T("kyc_get_failed"), res.Error)
		}
	}

	var kycDocs []model.KycDocument
	kycDocCond := model.KycDocument{
		KycDocument: ploutos.KycDocument{
			KycId: kyc.ID,
		},
	}
	res = model.DB.Where(kycDocCond).Find(&kycDocs)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		util.GetLoggerEntry(c).Errorf("Find kyc docs error: %s", res.Error.Error())
		return serializer.DBErr(c, service, i18n.T("kyc_get_failed"), res.Error)
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: serializer.BuildKyc(c, kyc, kycDocs),
	}
}
