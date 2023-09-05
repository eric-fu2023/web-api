package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

// UserRegisterService 管理用户注册服务
type UserRegisterService struct {
	Username   string `form:"username" json:"username" binding:"required"`
	Password   string `form:"password" json:"password" binding:"required"`
	CurrencyId int64  `form:"currency_id" json:"currency_id" binding:"required,numeric"`
}

func (service *UserRegisterService) Register(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	var existing model.User
	if r := model.DB.Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
		return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("密码加密失败"), err)
	}

	user := model.User{
		UserC: models.UserC{
			Username:   service.Username,
			Password:   string(bytes),
			Status:     1,
			Role:       1, // default role user
			CurrencyId: service.CurrencyId,
			BrandId:    int64(c.MustGet("_brand").(int)),
			AgentId:    int64(c.MustGet("_agent").(int)),
		},
	}

	err = CreateUser(user)
	if err != nil && errors.Is(err, ErrEmptyCurrencyId) {
		return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
	} else if err != nil && errors.Is(err, ErrFbCreateUserFailed) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
	} else if err != nil {
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
