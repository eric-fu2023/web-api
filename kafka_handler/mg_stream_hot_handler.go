package kafka_handler

import (
	"encoding/json"
	"os"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/IBM/sarama"
)

const ConsumerGroupIdMgStreamHot = "rf_stream_hot_getter"

type MgStreamHotHandler struct {
	Topic   string
	GroupId string
}

func NewMgStreamHotHandler(topic string) *MgStreamHotHandler {
	return &MgStreamHotHandler{
		Topic:   topic,
		GroupId: ConsumerGroupIdMgStreamHot,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (d *MgStreamHotHandler) Setup(session sarama.ConsumerGroupSession) error {
	util.Log().Info(d.GroupId + " has been set up")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (d *MgStreamHotHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the MgStreamHotHandler must finish its processing
// loop and exit.
func (d *MgStreamHotHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		msg, ok := <-claim.Messages()
		if !ok {
			util.Log().Error(d.GroupId + " message channel was closed")
			return nil
		}
		d.processMessages(msg)
		session.MarkMessage(msg, "")
	}
}

func (d *MgStreamHotHandler) processMessages(msg *sarama.ConsumerMessage) error {
	if msg.Topic != os.Getenv("MG_KAFKA_STREAM_HOT_TOPIC") {
		return errInvalidTopic
	}
	if len(msg.Value) == 0 {
		return errValueEmpty
	}
	var mgStreamHot ploutos.MgStreamHot
	err := json.Unmarshal(msg.Value, &mgStreamHot)
	if err != nil {
		return err
	}
	err = model.DB.Model(ploutos.LiveStream{}).Where(`mg_room_id`, mgStreamHot.RoomId).Update(`sort_factor`, mgStreamHot.Hot).Error
	if err != nil {
		return err
	}
	return nil
}
