package serializer

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"web-api/model"
)

type Profile struct {
	Nickname    string `json:"nickname"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	CountryCode string `json:"country_code"`
	Mobile      string `json:"mobile"`
	Avatar      string `json:"avatar"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name"`
	LastName    string `json:"last_name"`
	Street      string `json:"street"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Postcode    string `json:"postcode"`
	Birthday    string `json:"birthday"`
}

func BuildProfile(c *gin.Context, a model.UserProfile) (b Profile) {
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
	if !a.Birthday.IsZero() {
		b.Birthday = a.Birthday.Format(time.DateOnly)
	}
	if a.User != nil {
		b.Username = getMaskedUsername(a.User.Username)
		b.Email = getMaskedEmail(a.User.Email)
		b.CountryCode = a.User.CountryCode
		b.Mobile = getMaskedMobile(a.User.Mobile)
		b.Avatar = Url(a.User.Avatar)
	}
	return
}

func getMaskedUsername(original string) (new string) {
	l := len(original)
	if l == 0 {
		return
	}
	q := l / 3
	r := l % 3
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast + original[l-q:l]
	} else {
		new = original[:1] + "*"
	}
	return
}

func getMaskedEmail(original string) (new string) {
	if len(original) == 0 {
		return
	}
	l := strings.Index(original, "@")
	if l == -1 || l == 0 {
		new = original
		return
	}
	q := l / 2
	r := l % 2
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast
	} else {
		new = original[:1] + "*"
	}
	new += original[l:]
	return
}

func getMaskedMobile(original string) (new string) {
	l := len(original)
	if l == 0 {
		return
	}
	q := l / 2
	r := l % 2
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast
	} else {
		new = original[:1] + "*"
	}
	return
}
