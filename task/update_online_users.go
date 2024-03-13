package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"net/url"
	"os"
	"strings"
	"time"
	"web-api/cache"
)

var (
	Client         *autopaho.ConnectionManager
	connectedCh    chan []byte
	disconnectedCh chan []byte
)

type ConnectDisconnect struct {
	Username string `json:"username"`
	ClientId string `json:"clientid"`
}

func UpdateOnlineUsers() {
	connectedCh = make(chan []byte, 100)    // max 100 msg in buffer
	disconnectedCh = make(chan []byte, 100) // max 100 msg in buffer
	go func() {
		for {
			select {
			case msg, ok := <-connectedCh:
				if !ok { // channel closed
					return
				}
				var v ConnectDisconnect
				if e := json.Unmarshal(msg, &v); e == nil {
					if v.Username == "admin" {
						// do nothing
					} else if v.Username == "test" {
						cache.RedisSessionClient.SAdd(context.TODO(), "online_guests", v.ClientId)
					} else {
						cache.RedisSessionClient.SAdd(context.TODO(), "online_users", v.Username)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case msg, ok := <-disconnectedCh:
				if !ok { // channel closed
					return
				}
				var v ConnectDisconnect
				if e := json.Unmarshal(msg, &v); e == nil {
					if v.Username == "test" {
						cache.RedisSessionClient.SRem(context.TODO(), "online_guests", v.ClientId)
					} else {
						cache.RedisSessionClient.SRem(context.TODO(), "online_users", v.Username)
					}
				}
			}
		}
	}()
	Connect()
}

func Connect() {
	ctx := context.TODO()
	u, err := url.Parse(fmt.Sprintf("mqtts://%s", os.Getenv("MQTT_ADDRESS")))
	if err != nil {
		panic(err)
	}
	clientId := uuid.NewString()
	cfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{u},
		KeepAlive:         20,
		ConnectTimeout:    3 * time.Second,
		ConnectRetryDelay: 3 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: "$SYS/brokers/#", QoS: 1},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError: func(err error) { fmt.Printf("mqtt error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      clientId,
			OnClientError: func(err error) { fmt.Printf("mqtt connection error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("mqtt server error: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("mqtt server error; code: %d\n", d.ReasonCode)
				}
			},
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					//fmt.Printf("received message on topic %s; body: %s (retain: %t)\n", pr.Packet.Topic, pr.Packet.Payload, pr.Packet.Retain)
					if strings.Contains(pr.Packet.Topic, "/connected") {
						connectedCh <- pr.Packet.Payload
					} else if strings.Contains(pr.Packet.Topic, "/disconnected") {
						disconnectedCh <- pr.Packet.Payload
					}
					return true, nil
				}},
		},
	}
	cfg.SetUsernamePassword(os.Getenv("MQTT_USERNAME"), []byte(os.Getenv("MQTT_PASSWORD")))
	c, err := autopaho.NewConnection(ctx, cfg)
	if err != nil {
		panic(err)
	}
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}
	Client = c
	<-ctx.Done()
}
