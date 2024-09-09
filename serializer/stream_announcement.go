package serializer

import "web-api/model"

type StreamAnnouncement struct {
	AnnouncementText string `json:"announcement_text"`
}

func BuildStreamAnnouncement(saList []model.StreamAnnouncement) (output []StreamAnnouncement) {
	for _, sa := range saList {
		output = append(output, StreamAnnouncement{
			AnnouncementText: sa.AnnouncementDetail,
		})
	} 
	return
}

