package promotion

import (
	"context"
	"fmt"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"gorm.io/plugin/dbresolver"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
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

	if promotion.Type == ploutos.PromotionTypeDepositEarnMoreMission {

		missionTiers := GetPromotionMissionTiers(promotion.RewardDetails)

		if int(p.MissionId) >= len(missionTiers) {
			r = serializer.Err(c, p, serializer.CodeGeneralError, "Invalid Mission Tier", err)
			return
		}

		selectedMissionTierToClaim := missionTiers[p.MissionId]

		// 检查活动期间已充值金额
		topupRecords, getTopupsErr := model.TopupsByDateRange(c, user.ID, promotion.StartAt, promotion.EndAt)
		if getTopupsErr != nil {
			err = getTopupsErr
			return
		}
		totalDepositedAmount := util.Sum(topupRecords, func(co ploutos.CashOrder) int64 {
			return co.ActualCashInAmount
		})

		if totalDepositedAmount < selectedMissionTierToClaim.MissionAmount {
			r = serializer.Err(c, p, serializer.CodeGeneralError, "Deposit Amount Does Not Meet Requirement", err)
			return
		}

		// 检查该活动任务是否已完成过
		voucher, _ := model.GetVoucherByUserAndPromotionAndReference(c, user.ID, promotion.ID, fmt.Sprint(p.MissionId))
		if voucher.ID == 0 {

			coId := uuid.NewString()
			err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
				sum, err := model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "ValidateAndClaim"), tx, user.ID, selectedMissionTierToClaim.RewardAmount, selectedMissionTierToClaim.RewardAmount, 0, ploutos.TransactionTypeDepB, coId)
				if err != nil {
					return err
				}
				notes := fmt.Sprintf("UserId, PromotionId, MissionAmount, MissionRewardAmount - %v, %v, %v, %v", user.ID, promotion.ID, selectedMissionTierToClaim.MissionAmount, selectedMissionTierToClaim.RewardAmount)

				teamupCashOrder := ploutos.CashOrder{
					ID:                    coId,
					UserId:                user.ID,
					OrderType:             ploutos.CashOrderTypeDepB,
					Status:                ploutos.CashOrderStatusSuccess,
					Notes:                 ploutos.EncryptedStr(notes),
					AppliedCashInAmount:   selectedMissionTierToClaim.RewardAmount,
					ActualCashInAmount:    selectedMissionTierToClaim.RewardAmount,
					EffectiveCashInAmount: selectedMissionTierToClaim.RewardAmount,
					BalanceBefore:         sum.Balance - selectedMissionTierToClaim.RewardAmount,
					WagerChange:           selectedMissionTierToClaim.RewardAmount,
				}
				err = tx.Create(&teamupCashOrder).Error
				if err != nil {
					return
				}

				// Create a voucher
				voucher = ploutos.Voucher{
					UserID:             user.ID,
					Status:             ploutos.VoucherStatusRedeemed,
					Amount:             selectedMissionTierToClaim.RewardAmount,
					BrandID:            int64(brand),
					VoucherTemplateID:  0,
					ReferenceID:        fmt.Sprint(p.MissionId),
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

	if promotion.Type == ploutos.PromotionTypeDepositEarnMoreMission {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "Invalid Mission Id", err)
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
