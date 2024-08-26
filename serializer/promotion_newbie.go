package serializer

import "encoding/json"

type PromotionNewbieDetail struct {
	Title                  string                           `json:"title"`
	ImageUrl               string                           `json:"image_url"`
	EventWelcomePackage    PromotionNewbieWelcomePackage    `json:"event_welcome_package"`
	EventFirstDeposit      PromotionNewbieFirstDeposit      `json:"event_first_deposit"`
	EventRemittancePackage PromotionNewbieRemittancePackage `json:"event_remittance_package"`
	Footer                 []PromotionNewbieFooterContent   `json:"footer"`
}

type PromotionNewbieFooterContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// 新手大礼包
type PromotionNewbieWelcomePackage struct {
	EventId            int64  `json:"event_id"`
	Title              string `json:"title"`
	EndTime            int64  `json:"end_time"`
	ActionButtonStatus int64  `json:"action_button_status"`
	Description        string `json:"description"`
	IsActive bool `json:"is_active"`

	ButtonIconUrls   []string `json:"button_icon_urls"`
	IsPhoneBinded    bool     `json:"is_phone_binded"`
	IsBankcardBinded bool     `json:"is_bankcard_binded"`
}

// 存款初体验
type PromotionNewbieFirstDeposit struct {
	EventId            int64  `json:"event_id"`
	Title              string `json:"title"`
	EndTime            int64  `json:"end_time"`
	ActionButtonStatus int64  `json:"action_button_status"`
	Description        string `json:"description"`
	IsActive bool `json:"is_active"`

	ImageUrl string `json:"image_url"`
}

// 复活大礼包
type PromotionNewbieRemittancePackage struct {
	EventId            int64  `json:"event_id"`
	Title              string `json:"title"`
	EndTime            int64  `json:"end_time"`
	ActionButtonStatus int32  `json:"action_button_status"`
	Description        string `json:"description"`
	IsActive bool `json:"is_active"`

	ImageUrl        string  `json:"image_url"`
	NegativeRevenue float64 `json:"negative_revenue"`
}


func BuildNewbiePromotion() PromotionNewbieDetail{
	return PromotionNewbieDetail{}
}

func BuildDummyNewbiePromotion() PromotionNewbieDetail{
	var jsonData PromotionNewbieDetail
	err := json.Unmarshal([]byte(dummy), &jsonData)
	if err != nil {
		return PromotionNewbieDetail{}
	}
	return jsonData
}

var dummy = `
{
    "title": "新手MOCK",
      "image_url": "https://static.tayalive.com/aha-img/promotion_description/promotion_description-20240823091616-N1B8Yc.png",
      "event_welcome_package": {
        "event_id": 123,
        "title": "活动1:新手大礼包",
        "end_time": 1724979898000,
        "action_button_status": 1,
        "description": "奖励说明：自注册日起限时7日内完善绑定(手机号、绑定出款信息)可获10元彩金，彩金6倍流水出款。\n\n领奖期限：自注册日起，限时第7日23:59:59前完成绑定",
        "is_active": true,
        "button_icon_urls": [
            "https://static.tayalive.com/aha-img/promotion_image/promotion_image-20240823092409-W4kP0u.png",
            "https://static.tayalive.com/aha-img/promotion_description/promotion_description-20240823092407-S6nqXa.png",
            "https://static.tayalive.com/aha-img/promotion_description/promotion_description-20240823092406-Oset8I.png"
        ],
        "is_phone_binded": true,
        "is_bankcard_binded": false
      },
      "event_first_deposit": {
        "event_id": 124,
        "title": "活动2:存款初体验",
        "end_time": 1724979898000,
        "action_button_status": 1,
        "description": "奖励说明：自注册日起限时7日内完成单笔首存≥500元，即享彩金88至2888元，彩金6倍流水出款。\n\n领奖期限：自注册日起，限时第7日23:59:59前完成存款",
        "is_active": true,
        "image_url": "https://static.tayalive.com/aha-img/promotion_description/promotion_description-20240823091614-ZE0M4I.png"
      },
      "event_remittance_package": {
        "event_id": 125,
        "title": "活动3:复活大礼包",
        "end_time": 0,
        "action_button_status": 1,
        "description": "奖励说明：自首存日起限时7日内负盈利≥x元即可获得X%的复活礼金，最高X元。彩金6倍流水出款。\n\n领奖期限：申请截止第8天08:00，第9天23:59:59前派发至中心钱包",
        "is_active": true,
        "image_url": "https://static.tayalive.com/aha-img/promotion_image/promotion_image-20240823091615-gXNhH1.png",
        "negative_revenue": 0
      },
      "footer": [
        {
            "type": "title",
            "content": "参加新手三重奏"
        },
        {
            "type": "image",
            "content": "https://static.tayalive.com/aha-img/promotion_image/promotion_image-20240823091616-HSVBr2.png"
        },
        {
            "type": "title",
            "content": "活动规则"
        }
      ]
}
`