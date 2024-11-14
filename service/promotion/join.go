package promotion

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type PromotionJoin struct {
	PromotionId int64  `form:"promotion_id" json:"promotion_id"`
	Input       string `form:"input" json:"input"`
}

type PromotionJoinError struct {
	ErrorFields []int64 `json:"error_fields"`
}

func (p PromotionJoin) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()
	brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	// deviceInfo, _ := util.GetDeviceInfo(c)
	i18n := c.MustGet("i18n").(i18n.I18n)

	var errorFields []int64

	promotion, err := model.OngoingPromotionById(c, brand, p.PromotionId, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_not_found"), err)
		return
	}
	incomingRequestAction := serializer.IncomingPromotionRequestAction{}
	err = json.Unmarshal(promotion.Action, &incomingRequestAction)
	if err != nil {
		fmt.Println(err)
	}

	var requestInput map[string]string
	if p.Input == "" {
		p.Input = "{}"
	}
	err = json.Unmarshal([]byte(p.Input), &requestInput)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_fail"), err)
		return
	}

	isExceeded := false
	data := make(map[string]string)
	numOriFields := 0

	if len(incomingRequestAction.Fields) == 0 {
		parentPromotion, getParentPromotionErr := model.OngoingPromotionById(c, brand, promotion.ParentId, now)
		if getParentPromotionErr != nil {
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_not_found"), err)
			return
		}

		err = json.Unmarshal(parentPromotion.Action, &incomingRequestAction)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, field := range incomingRequestAction.Fields {
		if field.Switch == 0 {
			continue
		}
		switch field.Type {
		case "input_button":
			// numOriFields++
			isExceeded, err = serializer.ParseButtonClickOption(c, field, p.PromotionId, user.ID)
			if err != nil {
				r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_fail"), err)
				return
			}
		case "input_dropdown":
			numOriFields++
			if requestInput[field.InputId] != "" {
				index, _ := strconv.Atoi(requestInput[field.InputId])
				index--

				if index > len(field.Options)-1 {
					continue
				}
				option := field.Options[index]
				for _, value := range option {
					data[field.Title] = value
				}
			}
		case "input_keyin":
			numOriFields++
			if requestInput[field.InputId] != "" {
				data[field.Title] = requestInput[field.InputId]
			}
			contentTypeOption := field.ContentType
			switch int64(contentTypeOption) {
			case models.CustomPromotionTextboxOnlyInt:
				// If cast error means contains char
				_, castError := strconv.Atoi(requestInput[field.InputId])
				if castError != nil {
					r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_field_error"), err)
					return
				}
			}
		}
	}

	if numOriFields != len(data) {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_field_incomplete_error"), err)
		return
	}

	if isExceeded {
		r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("custom_promotion_entry_exceed"), err)
		return
	}

	jsonInput, _ := json.Marshal(data)

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

	if len(errorFields) > 0 {
		r = serializer.Response{
			Code: 50000,
			Data: PromotionJoinError{
				ErrorFields: errorFields,
			},
			Msg: i18n.T("custom_promotion_entry_field_error"),
		}
		return
	}
	r.Data = nil

	r = serializer.Response{
		Code: 0,
		Msg:  i18n.T("success"),
	}
	return
}
