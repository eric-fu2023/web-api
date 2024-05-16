package kafka_handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/IBM/sarama"
)

const ConsumerGroupIdMgStream = "rf_stream_getter"

var (
	errInvalidTopic = errors.New("invalid topic name")
	errValueEmpty   = errors.New("invalid msg value")
	errNotGame      = errors.New("data doesn't contain 'VS' and is not a game")
	errExclusion    = errors.New("nickname contains excluded keyword")
)

var sportsCategoryTypeMapping = map[int64]int64{
	1: 1, // sports
	3: 1, // sports
}

var sportsCategoryMapping = map[int64]int64{
	1: 1, // football
	3: 2, // basketball
}

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
	for _, s := range []string{"红队", "蓝队", "奇异果"} {
		if strings.Contains(mgStream.UserNickname, s) {
			return errExclusion
		}
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
	var stream ploutos.LiveStream
	err = model.DB.Where(`mg_room_id`, mgStream.RoomId).Find(&stream).Error
	if err != nil {
		return err
	}
	stream.Title = mgStream.Title
	stream.StreamerId = streamer.ID
	stream.ImgUrl = mgStream.Thumb
	stream.MgRoomId = &mgStream.RoomId
	stream.ScheduleTime = time.Now()
	if stream.ID == 0 {
		stream.Status = 1 // default pending
	}
	if mgStream.Srctp == 8 { // MatchId is FB id
		var match ploutos.Match
		if e := model.DB.Where(`match_id`, mgStream.MatchId).First(&match).Error; e == nil {
			stream.MatchId = match.ID
			stream.StreamCategoryTypeId = sportsCategoryTypeMapping[match.SportId]
			stream.StreamCategoryId = sportsCategoryMapping[match.SportId]
		}
	} else {
		//stream.MatchId, stream.StreamCategoryTypeId, stream.StreamCategoryId = SearchFBMatch(mgStream.Title, mgStream.League)
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
	err = model.DB.Save(&stream).Error
	if err != nil {
		return err
	}
	if mgStream.OfflineTime != 0 && stream.ID != 0 {
		go CallEndLiveApi(stream.ID)
	}
	return nil
}

func SearchFBMatch(title, league string) (int64, int64, int64) {
	title = strings.ToUpper(title)
	teams := strings.Split(title, "VS")
	if len(teams) != 2 {
		return 0, 0, 0
	}
	home := strings.TrimSpace(teams[0])
	away := strings.TrimSpace(teams[1])
	league = strings.TrimSpace(league)
	var match ploutos.Match
	err := model.DB.Where(`home_name_cn`, home).Where(`away_name_cn`, away).Where(`league_name_cn`, league).Order(`open_time DESC`).First(&match).Error
	if err != nil {
		return 0, 0, 0
	}
	fmt.Println("Mapping with FB match successful")
	return match.ID, sportsCategoryTypeMapping[match.SportId], sportsCategoryMapping[match.SportId]
}

func CallEndLiveApi(id int64) (err error) {
	apiUrl := os.Getenv("BACKEND_INTERNAL_BASE_URL") + "/internal/liveStream/endLiveStream"
	params, err := json.Marshal(map[string]int64{
		"id": id,
	})
	if err != nil {
		return
	}
	request, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(params))
	request.Header = http.Header{
		"Content-Type": []string{"application/json; charset=UTF-8"},
	}
	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return
	}
	return
}
