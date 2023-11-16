package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"web-api/conf/consts"
	"web-api/util"
)

type AuthEvent struct {
	ploutos.AuthEvent
}

func LogAuthEvent(event AuthEvent) error {
	coll := MongoDB.Collection(event.CollectionName())
	if _, err := coll.InsertOne(context.TODO(), event.AuthEvent); err != nil {
		return err
	}
	return nil
}

// GetLatestAuthEvents returns the latest successful login and failed password login events by the user
func GetLatestAuthEvents(userId int64, limit int) ([]AuthEvent, error) {
	// userid && (login && successful || login && failed && password)
	matchStage := bson.D{{
		Key: "$match", Value: bson.M{"$and": bson.A{
			bson.M{"userId": userId},
			bson.M{"$or": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"type": consts.AuthEventType["login"]},
					bson.M{"status": consts.AuthEventStatus["successful"]},
				}},
				bson.M{"$and": bson.A{
					bson.M{"type": consts.AuthEventType["login"]},
					bson.M{"status": consts.AuthEventStatus["failed"]},
					bson.M{"loginMethod": consts.AuthEventLoginMethod["password"]},
				}},
			}},
		}},
	}}
	sortStage := bson.D{{"$sort", bson.D{{"datetime", -1}}}}
	limitStage := bson.D{{"$limit", limit}}
	pipeline := mongo.Pipeline{
		matchStage,
		sortStage,
		limitStage,
	}

	coll := MongoDB.Collection(AuthEvent{}.CollectionName())
	allowDiskUseTrue := true
	opts := &options.AggregateOptions{AllowDiskUse: &allowDiskUseTrue}

	cursor, err := coll.Aggregate(context.Background(), pipeline, opts)
	if err != nil {
		util.Log().Error("Aggregate err", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var events []AuthEvent
	for cursor.Next(context.Background()) {
		curDoc := &ploutos.AuthEvent{}
		err = cursor.Decode(curDoc)
		if err != nil {
			util.Log().Error("Decode err", err)
			return nil, err
		}
		events = append(events, AuthEvent{AuthEvent: *curDoc})
	}

	return events, nil
}
