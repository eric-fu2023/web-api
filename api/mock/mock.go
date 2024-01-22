package mock

import (
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

func MockPromotonList(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Data: serializer.PromotionMock,
	})
}

func MockPromotonDetail(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Data: serializer.PromotionDetailMock,
	})
}

func MockOK(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Data: "success",
	})
}

func MockVoucherList(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Data: []serializer.Voucher{
			serializer.VoucherMock,
			serializer.VoucherMock2,
		},
	})
}

func MockVoucher(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Data: serializer.VoucherMock,
	})
}
