package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type StreamerGallery struct {
	ID         int64  `json:"id"`
	StreamerId int64  `json:"streamer_id"`
	Type       int64  `json:"type"`
	Src        string `json:"src"`
}

func BuildStreamerGallery(a ploutos.StreamerGallery) (b StreamerGallery) {
	b = StreamerGallery{
		ID:         a.ID,
		StreamerId: a.StreamerId,
		Type:       a.Type,
	}
	if a.Src != "" {
		b.Src = Url(a.Src)
	}
	return
}
