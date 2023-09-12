package model

type UserGallery struct {
	Base
	UserId	int64
	Type	int64
	Src		string
}

func (UserGallery) TableName() string {
	return "user_gallery"
}