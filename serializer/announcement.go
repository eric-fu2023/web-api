package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"web-api/conf/consts"
)

type Announcements struct {
	Texts    []string            `json:"texts,omitempty"`
	Images   []ImageAnnouncement `json:"images,omitempty"`
	Audios   []AudioAnnouncement `json:"audios,omitempty"`
	Overlays []string            `json:"overlays,omitempty"`
}

type ImageAnnouncement struct {
	Image string `json:"image"`
	Url   string `json:"url"`
}

type AudioAnnouncement struct {
	Text  string                 `json:"text,omitempty"`
	Image string                 `json:"image,omitempty"`
	Url   string                 `json:"url,omitempty"`
	Data  map[string]interface{} `json:"data"`
}

func BuildAnnouncements(a []ploutos.Announcement) (b Announcements) {
	for _, announcement := range a {
		if announcement.Type == consts.AnnouncementType["text"] {
			b.Texts = append(b.Texts, announcement.Text)
		} else if announcement.Type == consts.AnnouncementType["image"] {
			b.Images = append(b.Images, BuildImageAnnouncement(announcement))
		} else if announcement.Type == consts.AnnouncementType["audio"] {
			b.Audios = append(b.Audios, BuildAudioAnnouncement(announcement))
		} else if announcement.Type == consts.AnnouncementType["overlay"] {
			b.Overlays = append(b.Overlays, announcement.Text)
		}
	}
	return
}

func BuildImageAnnouncement(a ploutos.Announcement) (b ImageAnnouncement) {
	b = ImageAnnouncement{
		Image: Url(a.Image),
		Url:   a.Url,
	}
	return
}

func BuildAudioAnnouncement(a ploutos.Announcement) (b AudioAnnouncement) {
	b = AudioAnnouncement{
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
