package cashin_finpay_test

import (
	"net/http/httptest"
	"testing"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func TestCrypto(t *testing.T) {
	godotenv.Load("../../../.env")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	data, err := finpay.FinpayClient{}.PlaceDefaultCoinPalOrderV1(c, 20000, 1, "testestest1", "")
	t.Log(err)
	t.Log(data)
	t.Error(1)
}
