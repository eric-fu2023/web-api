package promotion

import (
	"fmt"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"gorm.io/plugin/dbresolver"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	userPromotionSessionClaimKey = "user_promotion_session_claim_lock:%d:%d"
)

type PromotionClaim struct {
	ID        int64 `form:"id" json:"id"`
	MissionId int64 `form:"mission_id" json:"mission_id"`
}

func (p PromotionClaim) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()
	brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)
	i18n := c.MustGet("i18n").(i18n.I18n)

	promotion, err := model.OngoingPromotionById(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}

	if p.MissionId != 0 {
		mission, _ := model.GetMissionById(c, brand, p.MissionId)
		if mission.ID == 0 {
			if err != nil {
				r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
				return
			}
		}
		topupRecords, getTopupsErr := model.TopupsByDateRange(c, user.ID, promotion.StartAt, promotion.EndAt)
		if getTopupsErr != nil {
			err = getTopupsErr
			return
		}
		totalDepositedAmount := util.Sum(topupRecords, func(co model.CashOrder) int64 {
			return co.ActualCashInAmount
		})

		if totalDepositedAmount < mission.MissionAmount {
			r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
			return
		}

		voucher, _ := model.GetVoucherByUserAndPromotionAndReference(c, user.ID, promotion.ID, p.MissionId)
		if voucher.ID == 0 {

			coId := uuid.NewString()
			err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
				sum, err := model.UpdateDbUserSumAndCreateTransaction(tx, user.ID, mission.RewardAmount, mission.RewardAmount, 0, ploutos.TransactionTypeDepB, coId)
				if err != nil {
					return err
				}
				notes := fmt.Sprintf("UserId, PromotionId, MissionId - %v, %v, %v", user.ID, promotion.ID, mission.ID)

				teamupCashOrder := ploutos.CashOrder{
					ID:                    coId,
					UserId:                user.ID,
					OrderType:             ploutos.CashOrderTypeDepB,
					Status:                ploutos.CashOrderStatusSuccess,
					Notes:                 ploutos.EncryptedStr(notes),
					AppliedCashInAmount:   mission.RewardAmount,
					ActualCashInAmount:    mission.RewardAmount,
					EffectiveCashInAmount: mission.RewardAmount,
					BalanceBefore:         sum.Balance - mission.RewardAmount,
					WagerChange:           mission.RewardAmount,
				}
				err = tx.Create(&teamupCashOrder).Error
				if err != nil {
					return
				}

				// Create a voucher
				voucher = ploutos.Voucher{
					UserID:             user.ID,
					Status:             ploutos.VoucherStatusRedeemed,
					Amount:             mission.RewardAmount,
					BrandID:            int64(brand),
					VoucherTemplateID:  0,
					ReferenceID:        fmt.Sprint(mission.ID),
					WagerMultiplier:    1,
					PromotionType:      promotion.Type,
					PromotionID:        promotion.ID,
					Name:               promotion.Name,
					PromotionSessionID: 0,
					UniqueID:           fmt.Sprint(time.Now().Unix()) + fmt.Sprint(user.ID),
				}
				err = tx.Save(&voucher).Error

				return
			})
		}

		r.Data = serializer.BuildVoucher(voucher, deviceInfo.Platform)
		return
	}

	// tz := time.FixedZone("local", int(promotion.Timezone))
	// now = now.In(tz)
	session, err := model.GetActivePromotionSessionByPromotionId(c, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	voucher, err := Claim(c, now, promotion, session, user.ID, &user)
	if err != nil {
		switch err.Error() {
		case "double_claim":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("double_claim"), err)
		case "unavailable_for_now":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("unavailable_for_now"), err)
		case "nothing_to_claim":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("nothing_to_claim"), err)
		default:
			r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		}
		return
	}
	r.Data = serializer.BuildVoucher(voucher, deviceInfo.Platform)
	return
}
