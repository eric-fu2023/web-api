package exchange_test

import (
	"context"
	"testing"
	"web-api/service/exchange"

	"github.com/joho/godotenv"
)

func TestCrypto(t *testing.T) {
	godotenv.Load("../../.env")

	data, err := exchange.OkxClient{}.GetExchangeRate(context.Background(), "", "")
	t.Log(err)
	t.Log(data)
	t.Error(1)
}
