package serializer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	model "web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type IncomingPromotion struct {
	Id     int64                   `json:"id"`
	Name   string                  `json:"name"`
	Images IncomingPromotionImages `json:"images"`
}

type IncomingPromotionImages struct {
	H5  string `json:"h5"`
	Web string `json:"web"`
}

type IncomingPromotionMatchList struct {
	List     []IncomingPromotionMatchListItem `json:"list"`
	Title    string                           `json:"title"`
	ListType string                           `json:"list_type"`
	Desc     []CustomPromotionPageDesc        `json:"desc"`
}

type IncomingPromotionMatchListItem struct {
	Id           string                               `json:"id"`
	Teams        []IncomingPromotionMatchListItemTeam `json:"teams"`
	Title        string                               `json:"title"`
	RedirectType string                               `json:"redirect_type"`
	Img          string                               `json:"img"`
	Name         string                               `json:"name"`
	Time         time.Time                            `json:"time"`
}

type IncomingPromotionMatchListItemTeam struct {
	HomeName string `json:"home_name"`
	HomeImg  string `json:"home_img"`
	AwayName string `json:"away_name"`
	AwayImg  string `json:"away_img"`
}

type IncomingPromotionRequestAction struct {
	Title       string                                `json:"title"`
	IsSubmitted bool                                  `json:"is_submitted"`
	Fields      []IncomingCustomPromotionRequestField `json:"fields"`
}

type IncomingCustomPromotionRequestField struct {
	Hint         string              `json:"hint"`
	Type         string              `json:"type"`
	Title        string              `json:"title"`
	InputId      int                 `json:"input_id"`
	Switch       int                 `json:"switch"`
	Options      []map[string]string `json:"option"`
	X            string              `json:"x"`
	Weightage    int                 `json:"weightage"`
	ErrorHint    string              `json:"error_hint,omitempty"`
	OrderType    string              `json:"order_type"`
	ContentType  string              `json:"content_type"`
	OrderStatus  string              `json:"order_status"`
	MaxClick     string              `json:"max_click"`
	RedirectType int                 `json:"redirect_type"`
}

type OutgoingCustomPromotionDetail struct {
	IsCustomPromotion  bool                             `json:"is_custom"`
	ParentInfo         IncomingPromotion                `json:"parent_info"`
	PromotionInfo      CustomPromotionPage              `json:"promotion_info,omitempty"`
	ChildrenPromotions []OutgoingCustomPromotionPreview `json:"children_promotions"`
}

type OutgoingCustomPromotionPreview struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type CustomPromotionDetail struct {
	Pages []CustomPromotionPage `json:"pages"`
}

type CustomPromotionPage struct {
	Title            string                    `json:"title"`
	PromotionId      int64                     `json:"promotion_id"`
	LoginStatus      int64                     `json:"login_status"`
	PageItemListType int64                     `json:"list_type"`
	PageItemList     []CustomPromotionPageItem `json:"list"`
	Desc             []CustomPromotionPageDesc `json:"desc"`
	Action           CustomPromotionRequest    `json:"action"`
}

type CustomPromotionPageItem struct {
	PageItemId    int64                     `json:"id"`
	StartDateTime time.Time                 `json:"start_date_time"`
	Name          string                    `json:"name"`
	Title         string                    `json:"title"`
	Teams         []CustomPromotionItemTeam `json:"teams"`
	Img           string                    `json:"img"`
	Text          string                    `json:"text"`
	// Input         CustomPromotionItemInput  `json:"input"`
	RedirectType int `json:"redirect_type"`
}

type CustomPromotionItemTeam struct {
	Name   string `json:"name"`
	ImgUrl string `json:"img"`
}
type CustomPromotionItemInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Url  string `json:"url"`
}

type CustomPromotionPageDesc struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type CustomPromotionRequest struct {
	Title        string                        `json:"title"`
	IsSubmitted  bool                          `json:"is_submitted"`
	Fields       []CustomPromotionRequestField `json:"fields"`
	RedirectType int64                         `json:"redirect_type"`
}

type CustomPromotionRequestField struct {
	Id          int    `json:"id"`
	Key         string `json:"key,omitempty"`
	Title       string `json:"title"`
	Placeholder string `json:"placeholder"`
	// Label       string                           `json:"label"`
	Type        string                           `json:"type"`
	Weightage   int                              `json:"weightage,omitempty"`
	Text        string                           `json:"text,omitempty"`
	Options     []CustomPromotionRequestDropdown `json:"options"`
	IntegerOnly bool                             `json:"integer_only"`
	ErrorMsg    string                           `json:"error_msg"`
	// RedirectType int                              `json:"redirect_type"`
}

type CustomPromotionRequestDropdown struct {
	Key   int    `json:"key"`
	Label string `json:"label"`
}

func BuildCustomPromotionDetail(progress, reward int64, platform string, p models.Promotion, s models.PromotionSession, v Voucher, cl ClaimStatus, extra any) PromotionDetail {
	raw := json.RawMessage(p.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}

	return PromotionDetail{
		ID:                     p.ID,
		Name:                   p.Name,
		Description:            json.RawMessage(p.Description),
		Image:                  Url(image),
		StartAt:                p.StartAt.Unix(),
		EndAt:                  p.EndAt.Unix(),
		ResetAt:                s.EndAt.Unix(),
		Type:                   int64(p.Type),
		RewardType:             int64(p.RewardType),
		RecurringDay:           int64(p.RecurringDay),
		RewardDistributionType: int64(p.RewardDistributionType),
		ClaimStatus:            cl,
		PromotionProgress:      BuildPromotionProgress(progress, p.GetRewardDetails()),
		Reward:                 float64(reward) / 100,
		Voucher:                v,
		Category:               int64(p.Category),
		IsVipAssociated:        p.VipAssociated,
		DisplayOnly:            p.DisplayOnly,
		Extra:                  extra,
	}
}

func BuildPromotionMatchList(incoming []IncomingPromotionMatchListItem, subPromotion models.Promotion) (res []CustomPromotionPageItem) {

	for _, item := range incoming {
		id, _ := strconv.Atoi(item.Id)
		outgoingPageItem := CustomPromotionPageItem{
			PageItemId: int64(id),
			// Type:       item.Type,
			Title:         subPromotion.Name,
			Text:          "立即前往",
			StartDateTime: item.Time,
		}

		if item.Title != "" {
			outgoingPageItem.Title = item.Title
			outgoingPageItem.Name = item.Name
			outgoingPageItem.Img = item.Img
		}

		outgoingTeams := []CustomPromotionItemTeam{}
		for _, team := range item.Teams {
			teamItem := CustomPromotionItemTeam{}
			if team.HomeName == "" && team.HomeImg == "" {
				teamItem.Name = team.AwayName
				teamItem.ImgUrl = team.AwayImg
			}
			if team.AwayName == "" && team.AwayImg == "" {
				teamItem.Name = team.HomeName
				teamItem.ImgUrl = team.HomeImg
			}

			outgoingTeams = append(outgoingTeams, teamItem)
		}

		outgoingPageItem.Teams = outgoingTeams

		res = append(res, outgoingPageItem)
	}

	return
}

func BuildPromotionAction(c *gin.Context, incoming IncomingPromotionRequestAction, promotionId int64, userId int64, loginStatusType int64) (res CustomPromotionRequest) {

	res.Title = incoming.Title

	if userId == 0 && loginStatusType == models.CustomPromotionLoginStatusLogin {
		res.Title = "立即参与，享受专属福利！"
		res.RedirectType = models.CustomPromotionButtonRedirectTypeLogin
		res.Fields = append(res.Fields, CustomPromotionRequestField{
			Title: "立即登录，抢先参与",
			Text:  "立即登录，抢先参与",
			Type:  "button",
		})

		return
	}

	for _, incomingField := range incoming.Fields {
		if incomingField.Switch == 0 {
			continue
		}
		requestField := CustomPromotionRequestField{
			Placeholder: incomingField.Hint,
			Title:       incomingField.Title,
			Type:        strings.Replace(incomingField.Type, "input_", "", -1),
			Id:          incomingField.InputId,
			Weightage:   incomingField.Weightage,
			ErrorMsg:    incomingField.ErrorHint,
		}

		if requestField.Type == "button" {
			requestField.Text = requestField.Title

			if incomingField.RedirectType == 0 {
				isExceeded, err := ParseButtonClickOption(c, incomingField, promotionId, userId)
				if err != nil {
					fmt.Println(err)
				}

				res.IsSubmitted = isExceeded
				if res.IsSubmitted {
					res.Title = "感谢您的参与！"
				}
			} else {
				// requestField.RedirectType = incomingField.RedirectType
				res.RedirectType = int64(incomingField.RedirectType)
			}
		}

		if requestField.Type == "dropdown" {
			keyIndex := 1
			for _, option := range incomingField.Options {
				for _, value := range option {

					requestField.Options = append(requestField.Options, CustomPromotionRequestDropdown{
						Key:   keyIndex,
						Label: value,
					})
				}
				keyIndex++
			}
		} else {
			requestField.Options = make([]CustomPromotionRequestDropdown, 0)
		}

		if requestField.Type == "keyin" {
			contentTypeOption, _ := strconv.Atoi(incomingField.ContentType)
			switch int64(contentTypeOption) {
			case models.CustomPromotionTextboxOnlyInt:
				requestField.IntegerOnly = true
			}
		}

		res.Fields = append(res.Fields, requestField)
	}

	return
}

func ParseButtonClickOption(c *gin.Context, incoming IncomingCustomPromotionRequestField, promotionId, userId int64) (isExceeded bool, err error) {

	buttonClickOption, _ := strconv.Atoi(incoming.MaxClick)
	entryLimitType := int64(buttonClickOption)
	if incoming.X == "" {
		incoming.X = "0"
	}
	buttonClickTimes, _ := strconv.Atoi(incoming.X)

	isExceeded, err = model.CheckIfCustomPromotionEntryExceededLimit(c, entryLimitType, promotionId, userId, buttonClickTimes)
	if err != nil {
		fmt.Println(err)
	}

	return
}
