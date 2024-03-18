package task

const (
	RedisKeyOnlineUsers      = "online_users"
	RedisKeyOnlineGuests     = "online_guests"
	RedisKeySubscribedUsers  = "subscribed:%s:users"
	RedisKeySubscribedGuests = "subscribed:%s:guests"
)

type BaseMQTTMsg struct {
	Username string `json:"username"`
	ClientId string `json:"clientid"`
}

type ConnectDisconnect struct {
	BaseMQTTMsg
}

type SubscribeUnsubscribe struct {
	BaseMQTTMsg
	Topic string `json:"topic"`
}
