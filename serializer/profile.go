package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"time"
)

type Profile struct {
	Nickname   string `json:"nickname"`
	Pic        string `json:"pic"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Street     string `json:"street"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Postcode   string `json:"postcode"`
	Birthday   string `json:"birthday"`
}

func BuildProfile(c *gin.Context, a ploutos.UserProfile) (b Profile) {
	b = Profile{
		Nickname:   a.Nickname,
		FirstName:  a.FirstName,
		MiddleName: a.MiddleName,
		LastName:   a.LastName,
		Street:     a.Street,
		Province:   a.Province,
		City:       a.City,
		Postcode:   a.Postcode,
	}
	if a.Pic != "" {
		b.Pic = Url(a.Pic)
	}
	if !a.Birthday.IsZero() {
		b.Birthday = a.Birthday.Format(time.DateOnly)
	}
	return
}
