package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type ImageAnnouncement struct {
	Image string `json:"image"`
	Url   string `json:"url"`
}

func BuildImageAnnouncement(a ploutos.Announcement) (b ImageAnnouncement) {
	b = ImageAnnouncement{
		Image: Url(a.Image),
		Url:   a.Url,
	}
	return
}
