package referral_alliance

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
)

type RankingsService struct{}

func (s *RankingsService) Get(c *gin.Context) (r serializer.Response, err error) {
	return
}
