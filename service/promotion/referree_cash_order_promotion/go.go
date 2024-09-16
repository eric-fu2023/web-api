package referree_cash_order_promotion

//
//import (
//	po "blgit.rfdev.tech/taya/ploutos-object"
//	"context"
//	"errors"
//
//	"gorm.io/gorm"
//)
//
//type Service struct {
//	db *gorm.DB
//}
//
//func NewService(db *gorm.DB) (*Service, error) {
//	if db == nil {
//		return nil, errors.New("no gorm")
//	}
//	return &Service{}, nil
//}
//
//func (s *Service) AddReward(ctx context.Context, referree UserForm, cashOrderInfo int64) error {
//	var referreeDbInfo po.User
//	db := s.db.Where(`username`, referree.Id)
//	if err := db.Scopes(po.ByActiveNonStreamerUser).First(&referreeDbInfo).Error; err != nil {
//		return err
//	}
//
//	tx := db.Debug().Table("user_referrals").Joins("LEFT JOIN users ON users.id = user_referrals.referral_id").Where("user_referrals.referral_id = ?", referree.Id).Select("game_integrations.id as giid, game_vendor.id as gvid").Find(&ref)
//
//	return nil
//}
//
//type CashOrder struct {
//	po.CashOrder
//}
