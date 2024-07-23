package promotion

import (
	"encoding/json"
	"fmt"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type PromotionJoin struct {
	PromotionId int64                  `form:"promotion_id" json:"promotion_id"`
	Input       map[string]interface{} `form:"input" json:"input"`
}

func (p PromotionJoin) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()
	brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	// deviceInfo, _ := util.GetDeviceInfo(c)
	i18n := c.MustGet("i18n").(i18n.I18n)

	promotion, err := model.PromotionGetActive(c, brand, p.PromotionId, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_not_found"), err)
		return
	}
	incomingRequestAction := serializer.IncomingPromotionRequestAction{}
	err = json.Unmarshal(promotion.Action, &incomingRequestAction)
	if err != nil {
		fmt.Println(err)
	}

	isExceeded := false
	for _, field := range incomingRequestAction.Fields {
		if field.Type == "input-button" {
			isExceeded, err = serializer.ParseButtonClickOption(c, field, p.PromotionId, user.ID)
			if err != nil {
				r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_fail"), err)
				return
			}
		}
	}

	if isExceeded {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_exceed"), err)
		return
	}

	jsonInput, _ := json.Marshal(p.Input)

	request := models.PromotionRequest{
		PromotionId:  p.PromotionId,
		UserId:       user.ID,
		Status:       1, // Pending
		InputDetails: jsonInput,
	}

	err = model.CreateJoinCustomPromotion(request)

	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_fail"), err)
		return
	}

	r.Data = nil

	return
}
