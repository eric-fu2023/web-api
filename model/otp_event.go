package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"context"
)

type OtpEvent struct {
	models.OtpEventC
}

func LogOtpEvent(event OtpEvent) error {
	coll := MongoDB.Collection(event.CollectionName())
	if _, err := coll.InsertOne(context.TODO(), event.OtpEventC); err != nil {
		return err
	}
	return nil
}
