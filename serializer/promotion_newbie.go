package serializer

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
