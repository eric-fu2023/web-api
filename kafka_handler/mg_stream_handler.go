package kafka_handler

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"gorm.io/gorm/clause"
	"os"
	"strconv"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

const ConsumerGroupIdMgStream = "rf_stream_getter"

var (
	errInvalidTopic = errors.New("invalid topic name")
	errValueEmpty   = errors.New("invalid msg value")
	errNotGame      = errors.New("data doesn't contain 'VS' and is not a game")
)

type MgStreamHandler struct {
	Topic   string
	GroupId string
}

func NewMgStreamHandler(topic string) *MgStreamHandler {
	intervalSeconds, err := strconv.Atoi(os.Getenv("DATA_PIPELINE_SAVE_USER_ACTIVITY_LOGS_INTERVAL_SECONDS"))
	if intervalSeconds == 0 {
		util.Log().Error("Err parsing SaveUserActivityLogsBatchSize", err.Error())
		intervalSeconds = 1
	}
	return &MgStreamHandler{
		Topic:   topic,
		GroupId: ConsumerGroupIdMgStream,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (d *MgStreamHandler) Setup(session sarama.ConsumerGroupSession) error {
	util.Log().Info(d.GroupId + " has been set up")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (d *MgStreamHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the MgStreamHandler must finish its processing
// loop and exit.
func (d *MgStreamHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

func (d *MgStreamHandler) processMessages(msg *sarama.ConsumerMessage) error {
	if msg.Topic != os.Getenv("MG_KAFKA_STREAM_TOPIC") {
		return errInvalidTopic
	}
	if len(msg.Value) == 0 {
		return errValueEmpty
	}
	var mgStream ploutos.MgStream
	err := json.Unmarshal(msg.Value, &mgStream)
	if err != nil {
		return err
	}
	if !strings.Contains(mgStream.Title, "VS") {
		return errNotGame
	}
	var streamer ploutos.User
	err = model.DB.Where(`username`, fmt.Sprintf(`mg-streamer-%v`, mgStream.UserNickname)).Where(`role`, consts.UserRole["streamer"]).Find(&streamer).Error
	if err != nil {
		return err
	}
	if streamer.ID == 0 || streamer.Avatar != mgStream.UserAvatar {
		streamer.Username = fmt.Sprintf(`mg-streamer-%v`, mgStream.UserNickname)
		streamer.Role = consts.UserRole["streamer"]
		streamer.Avatar = mgStream.UserAvatar
		streamer.Nickname = mgStream.UserNickname
		err = model.DB.Save(&streamer).Error
		if err != nil {
			return err
		}
	}
	stream := ploutos.LiveStream{
		Title:        mgStream.Title,
		StreamerId:   streamer.ID,
		Status:       1, // default pending
		ImgUrl:       mgStream.Thumb,
		MgRoomId:     &mgStream.RoomId,
		ScheduleTime: time.Now(),
	}
	if mgStream.Srctp == 8 { // MatchId is FB id
		stream.MatchId = mgStream.MatchId
	} else {
		stream.MatchId = SearchFBMatch(mgStream.Title, mgStream.League)
	}
	if mgStream.OnlineTime != 0 {
		stream.Status = 2
		stream.OnlineAt = time.Unix(mgStream.OnlineTime, 0)
	}
	if mgStream.MediaStream != "" || mgStream.MediaFlvStream != "" {
		url := map[string]string{
			"flv":  "", // mgStream.MediaFlvStream
			"m3u8": mgStream.MediaStream,
		}
		if j, e := json.Marshal(&url); e == nil {
			stream.PullUrl = string(j)
		}
	}
	if mgStream.OfflineTime != 0 {
		stream.Status = 3
		stream.OfflineAt = time.Unix(mgStream.OfflineTime, 0)
	}
	err = model.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "mg_room_id"}},
		UpdateAll: true,
	}).Create(&stream).Error
	if err != nil {
		return err
	}
	return nil
}

func SearchFBMatch(title, league string) int64 {
	title = strings.ToUpper(title)
	teams := strings.Split(title, "VS")
	if len(teams) != 2 {
		return 0
	}
	home := strings.TrimSpace(teams[0])
	away := strings.TrimSpace(teams[1])
	league = strings.TrimSpace(league)
	var match ploutos.Match
	err := model.DB.Where(`home_name_cn`, home).Where(`away_name_cn`, away).Where(`league_name_cn`, league).Order(`open_time DESC`).First(&match).Error
	if err != nil {
		return 0
	}
	fmt.Println("Mapping with FB match successful")
	return match.MatchId
}
