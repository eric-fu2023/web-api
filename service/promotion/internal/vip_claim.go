package internal

type VipClaimRequest struct {
	UserID            int64 //needed
	VipLevel          int64 //for unique id annotation
	PromotionID       int64 //craft voucher
	VipRewardRecordID int64 //mark status
	Amount            int64 //calculation is done
}

// get voucher template
// craft voucher
// claim promotion - check unique id
// mark vip reward record
