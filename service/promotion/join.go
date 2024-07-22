package promotion

import (
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type PromotionJoin struct {
	ID     int64                  `form:"id" json:"id"`
	UserId int64                  `json:"user_id"`
	Input  []PromotionJoinRequest `form:"input" json:"input"`
}
type PromotionJoinRequest struct {
	InputKey   string `form:"input_key" json:"input_key"`
	InputValue string `form:"input_value" json:"input_value"`
}

func (p PromotionJoin) Handle(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now().UTC()
	// brand := c.MustGet(`_brand`).(int)
	// user := c.MustGet("user").(model.User)
	// deviceInfo, _ := util.GetDeviceInfo(c)
	// i18n := c.MustGet("i18n").(i18n.I18n)

	// p.UserId = user.ID

	// // Insert into DB

	// // Get Ploustos Object
	// joinEntry, _ := model.FindJoinCustomPromotionEntry(c, brand, p.ID)

	// // If joinEntry status is Rejected / Pending
	// if joinEntry.Id == 0 {
	// 	// request := parseIncomingPromotionJoin(p)
	// 	request := PromotionJoin{
	// 		ID: 81,
	// 	}
	// 	request.Input = append(request.Input, PromotionJoinRequest{})
	// 	data := make(map[string]interface{})

	// 	data["TestKey"] = "TestValue"
	// 	data["How are you?"] = "I'm Fine"
	// 	data["我谁?"] = "你Fine"

	// 	jsonData, _ := json.Marshal(data)
	// 	fmt.Println(string(jsonData))
	// 	joinRecord := model.CreateJoinCustomPromotion(request)
	// }

	// r.Data = nil

	return
}

// func parseIncomingPromotionJoin(p PromotionJoin) (request models.PromotionRequest) {

// }
