package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RoomMessage struct {
	Id        string `bson:"_id"`
	Room      string
	Message   string
	Nickname  string
	Avatar    string
	UserType  int64 `bson:"user_type"`
	Type      int64
	UserId    int64 `bson:"user_id"`
	Sender    *ploutos.User
	Timestamp int64
	VipId     int64 `bson:"vip_id"`
}

func (a RoomMessage) List(room string, from int64, page int64, limit int64) (r []RoomMessage, err error) {
	ctx := context.TODO()
	coll := MongoDB.Collection("room_message")
	filter := bson.M{"room": room, "deleted_at": nil}
	opts := options.Find()
	opts.SetLimit(limit)
	opts.SetSort(bson.D{{"timestamp", -1}, {"_id", -1}})
	if page > 0 {
		opts.SetSkip((page - 1) * limit)
	}
	if from != 0 {
		filter["timestamp"] = bson.M{"$gte": from}
	}
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return
	}
	for cursor.Next(ctx) {
		var pm RoomMessage
		cursor.Decode(&pm)
		r = append(r, pm)
	}
	return
}
