package referree_cash_order_promotion

import (
	"context"
)

type Service struct {
}

func NewService() (*Service, error) {
	return &Service{}, nil
}

func (s *Service) AddReward(ctx context.Context, referree UserForm, rewardId int64) error {

	return nil

}
