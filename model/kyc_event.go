package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
)

type KycEvent struct {
	ploutos.KycEvent
}

func LogKycEvent(event KycEvent) error {
	coll := MongoDB.Collection(event.CollectionName())
	if _, err := coll.InsertOne(context.TODO(), event.KycEvent); err != nil {
		return err
	}
	return nil
}
