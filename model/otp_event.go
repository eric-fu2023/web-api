package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
)

type OtpEvent struct {
	ploutos.OtpEvent
}

func LogOtpEvent(event OtpEvent) error {
	coll := MongoDB.Collection(event.CollectionName())
	if _, err := coll.InsertOne(context.TODO(), event.OtpEvent); err != nil {
		return err
	}
	return nil
}
