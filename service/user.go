package service

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/dc"
	"web-api/service/dollar_jackpot"
	"web-api/service/fb"
	"web-api/service/imsb"
	"web-api/service/saba"
	"web-api/service/stream_game"
	"web-api/service/taya"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var (
	ErrEmptyCurrencyId           = errors.New("empty currency id")
	GameVendorUserRegisterStruct = map[string]common.UserRegisterInterface{
		"taya":           &taya.UserRegister{},
		"fb":             &fb.UserRegister{},
		"saba":           &saba.UserRegister{},
		"dc":             &dc.UserRegister{},
		"imsb":           &imsb.UserRegister{},
		"dollar_jackpot": &dollar_jackpot.UserRegister{},
		"stream_game":    &stream_game.UserRegister{},
	}
)

var (
	ErrTokenGeneration = errors.New("token generation error")
)

func CreateNewUser(user *model.User, referralCode string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = CreateNewUserWithDB(user, referralCode, tx)
		return
	})
	return
}

func CreateNewUserWithDB(user *model.User, referralCode string, tx *gorm.DB) (err error) {

	// Default AgentId and ChannelId if no channelCode
	// Change to env later
	agentIdString := os.Getenv("DEFAULT_AGENT_ID")
	channelCode := ""

	if agentIdString == "" || agentIdString == "1000000" {
		agentIdString = "1000001"
	}

	var channelId int64

	agentId, err := strconv.Atoi(agentIdString)

	if err != nil {
		return fmt.Errorf("string conv err: %w", err)
	}

	// prefixExist := true

	// Check AgentCode and ChannelCode, COMBINATION OF BOTH MUST EXIST, else treat as default

	invalidAgentCode := false

	// 1) Check AgentCode
	// 2) Check ChannelCode

	splitChannel := strings.Split(user.Channel, "_")

	splittedAgentCode := strings.Join(splitChannel[:len(splitChannel)-1], "_")
	splittedChannelCode := splitChannel[len(splitChannel)-1]

	// Eg: C1000 (Not C1000_1, 代理_渠道号)
	if len(splitChannel) < 2 {
		splittedAgentCode = splittedChannelCode
		splittedChannelCode = ""
	}

	agent := ploutos.Agent{
		Code: splittedAgentCode,
	}

	err = tx.Where(`code`, splittedAgentCode).Find(&agent).Error
	if err != nil {
		return
	}

	if agent.ID == 0 {
		invalidAgentCode = true
	}

	if invalidAgentCode {
		user.Channel = ""
	} else {
		channel := ploutos.Channel{
			AgentId: int64(agent.ID),
			Code:    splittedChannelCode,
		}

		err = tx.Where(`agent_id`, agent.ID).Where(`code`, splittedChannelCode).Find(&channel).Error
		if err != nil {
			return
		}

		if splittedChannelCode == "" {
			if channel.ID == 0 {
				err = tx.Create(&channel).Error
				if err != nil {
					return
				}
			}
			channelId = channel.ID
		} else {
			if channel.ID == 0 {
				user.Channel = ""
			}
		}
	}

	if user.Channel != "" {

		agentCode := splittedAgentCode
		// channelCode = splittedChannelCode

		agent := ploutos.Agent{
			Code: agentCode,
		}

		err = tx.Where(`code`, agentCode).Find(&agent).Error
		if err != nil {
			return
		}

		if agent.ID == 0 {

			// Get Default instead of Create New
			agent.ID = int64(agentId)
			agentCode = ""

			// prefixExist = false
		}

		user.AgentId = agent.ID

		// if user.Channel = agent means user.Channel has no channelCode suffix
		// if user.Channel != agent means user.Channel has channelCode suffix
		user.Channel = agentCode
		user.ChannelId = channelId

	} else {
		user.AgentId = int64(agentId)
		channel := ploutos.Channel{
			Code:    splittedChannelCode,
			AgentId: int64(agentId),
		}

		err = tx.Where(`agent_id`, agentId).Where(`code`, channelCode).Find(&channel).Error
		if err != nil {
			return
		}
		if channel.ID == 0 {
			err = tx.Create(&channel).Error
			if err != nil {
				return
			}
		}
		user.ChannelId = channel.ID
	}

	var existed model.User
	rows := model.DB.Where(`username`, user.Username).First(&existed).RowsAffected
	if rows > 0 {
		return fmt.Errorf("username existed: %s %w", user.Username, err)
	}

	err = user.CreateWithDB(tx)
	if err != nil {
		return fmt.Errorf("create with db: %w", err)
	}

	if referralCode != "" {
		referrer, err := model.LinkReferralWithDB(tx, user.ID, referralCode)
		if err != nil {
			return fmt.Errorf("link referral with db: %w", err)
		}
		// Overwrite user's channel with referrer's channel
		// Set to empty if referrer's channel is empty
		user.Channel = referrer.Channel
		err = tx.Select("channel").Updates(&model.User{User: ploutos.User{
			BASE:    ploutos.BASE{ID: user.ID},
			Channel: user.Channel,
		}}).Error
		if err != nil {
			return fmt.Errorf("update user channel: %w", err)
		}
	}

	return nil
}

func CreateUser(user *model.User) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = CreateUserWithDB(user, tx)
		return
	})
	return
}

func CreateUserWithDB(user *model.User, tx *gorm.DB) (err error) {
	err = tx.Save(&user).Error
	if err != nil {
		return
	}

	userSum := ploutos.UserSum{
		UserId: user.ID,
	}
	tx2 := model.DB.Clauses(dbresolver.Use("txConn")).Begin()
	err = tx2.Error
	if err != nil {
		return
	}
	err = tx2.Create(&userSum).Error
	if err != nil {
		tx2.Rollback()
		return
	}

	userCounter := ploutos.UserCounter{
		UserId: user.ID,
	}
	err = tx.Create(&userCounter).Error
	if err != nil {
		tx2.Rollback()
		return
	}

	err = tx.Model(ploutos.User{}).Where(`id`, user.ID).Update(`setup_completed_at`, time.Now()).Error
	if err != nil {
		tx2.Rollback()
		return
	}

	var integrationCurrencies []ploutos.CurrencyGameIntegration
	err = tx.Where(`currency_id`, user.CurrencyId).Find(&integrationCurrencies).Error
	if err != nil {
		tx2.Rollback()
		return ErrEmptyCurrencyId
	}

	inteCurrMap := make(map[int64]string)
	for _, cur := range integrationCurrencies {
		inteCurrMap[cur.GameIntegrationId] = cur.Value
	}

	var gameIntegrations []ploutos.GameIntegration
	err = tx.Model(ploutos.GameIntegration{}).Find(&gameIntegrations).Error
	if err != nil {
		tx2.Rollback()
		return
	}
	for _, gi := range gameIntegrations {
		currency, exists := inteCurrMap[gi.ID]
		if !exists {
			tx2.Rollback()
			return ErrEmptyCurrencyId
		}

		integratedGame, found := common.GameIntegration[gi.ID]
		if !found {
			return errors.New(fmt.Sprintf("integrated game id %d not found", gi.ID))
		}

		err = integratedGame.CreateWallet(*user, currency)
		if err != nil {
			tx2.Rollback()
			return
		}
	}
	// TODO: might remove in the future
	var currencies []ploutos.CurrencyGameVendor
	err = tx.Where(`currency_id`, user.CurrencyId).Find(&currencies).Error
	if err != nil {
		tx2.Rollback()
		return ErrEmptyCurrencyId
	}

	currMap := make(map[int64]string)
	for _, cur := range currencies {
		currMap[cur.GameVendorId] = cur.Value
	}

	games := strings.Split(os.Getenv("GAMES_REGISTERED_FOR_NEW_USER"), ",")
	for _, g := range games {
		if g == "" {
			continue
		}
		currency, exists := currMap[consts.GameVendor[g]]
		if !exists {
			tx2.Rollback()
			return ErrEmptyCurrencyId
		}
		game := GameVendorUserRegisterStruct[g]
		e := game.CreateUser(*user, currency)
		if e != nil && !errors.Is(e, game.VendorRegisterError()) { // if create vendor user failed, can proceed safely. when user first enter the game, it will retry
			tx2.Rollback()
			return fmt.Errorf("%w: %w", game.OthersError(), e)
		}
	}
	// TODO: END

	tx2.Commit()
	return
}

type MeService struct {
	WithKyc bool `form:"with_kyc" json:"with_kyc"`
}

func (service *MeService) Get(c *gin.Context) serializer.Response {
	u, _ := c.Get("user")
	user := u.(model.User)
	var userSum ploutos.UserSum

	if !user.IsDeposited {
		firstTime, err := model.CashOrder{}.IsFirstTime(c, user.ID)
		// Not first time = deposited before
		if err == nil && !firstTime {
			user.IsDeposited = true
			_ = model.DB.Save(&user).Error
		}
	}

	if e := model.DB.Where(`user_id`, user.ID).First(&userSum).Error; e == nil {
		user.UserSum = &userSum
	}

	userAchievements, err := model.GetUserAchievementsForMe(user.ID)
	if err == nil {
		user.Achievements = userAchievements
	}

	if service.WithKyc {
		if value, err := GetCachedConfig(c, consts.ConfigKeyTopupKycCheck); err == nil {
			if value == "true" {
				user.KycCheckRequired = true
			}
		}
		var kyc model.Kyc
		if rows := model.DB.Scopes(model.ByUserId(user.ID), model.BySuccess).Order(`id DESC`).Find(&kyc).RowsAffected; rows > 0 {
			user.Kyc = &kyc
		}
	}
	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}
