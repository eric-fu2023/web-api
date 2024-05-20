package task

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"os"
	"strings"
	"web-api/kafka_handler"
	"web-api/util"
)

func ConsumeMgStreamsHot() error {
	conn := os.Getenv("MG_KAFKA_CONN")
	topic := os.Getenv("MG_KAFKA_STREAM_HOT_TOPIC")
	clientId := os.Getenv("MG_KAFKA_CLIENT_ID")
	if conn == "" || topic == "" || clientId == "" {
		return nil
	}
	conns := strings.Split(conn, ",")
	defaultConfig := sarama.NewConfig()
	defaultConfig.Metadata.AllowAutoTopicCreation = false
	defaultConfig.Version = sarama.V3_6_0_0
	defaultConfig.ClientID = clientId
	defaultConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerGroup, err := sarama.NewConsumerGroup(conns, kafka_handler.ConsumerGroupIdMgStreamHot, defaultConfig)
	if err != nil {
		util.Log().Error("Error NewConsumerGroup", err.Error())
		return err
	}

	handler := kafka_handler.NewMgStreamHotHandler(topic)
	ctx := context.Background()
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				util.Log().Panic("Error from consumer: %v", err) // Should not happen
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}
