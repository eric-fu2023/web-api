package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"web-api/conf/consts"
)

type Announcements struct {
	Texts    []string            `json:"texts,omitempty"`
	Images   []OtherAnnouncement `json:"images,omitempty"`
	Audios   []OtherAnnouncement `json:"audios,omitempty"`
	Overlays []OtherAnnouncement `json:"overlays,omitempty"`
}

type OtherAnnouncement struct {
	Text  string                 `json:"text,omitempty"`
	Image string                 `json:"image,omitempty"`
	Url   string                 `json:"url,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func BuildAnnouncements(a []ploutos.Announcement) (b Announcements) {
	for _, announcement := range a {
		if announcement.Type == consts.AnnouncementType["text"] {
			b.Texts = append(b.Texts, announcement.Text)
		} else if announcement.Type == consts.AnnouncementType["image"] {
			b.Images = append(b.Images, BuildOtherAnnouncement(announcement))
		} else if announcement.Type == consts.AnnouncementType["audio"] {
			b.Audios = append(b.Audios, BuildOtherAnnouncement(announcement))
		} else if announcement.Type == consts.AnnouncementType["overlay"] {
			b.Overlays = append(b.Overlays, BuildOtherAnnouncement(announcement))
		}
	}
	return
}

func BuildOtherAnnouncement(a ploutos.Announcement) (b OtherAnnouncement) {
	b = OtherAnnouncement{
		Text:  a.Text,
		Image: Url(a.Image),
		Url:   a.Url,
	}
	var data map[string]interface{}
	if e := json.Unmarshal(a.Data, &data); e == nil {
		b.Data = data
	}
	return
}
