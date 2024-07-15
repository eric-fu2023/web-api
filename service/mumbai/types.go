package mumbai

import (
	"errors"
)

var ErrInsufficientMumbaiWalletBalance = errors.New("insufficient mumbai wallet balance")
var ErrGetBalance = errors.New("Mumbai get balance error")
var ErrGameCodeMapping = errors.New("game code mapping error")

type ResponseCode string

const (
	ResponseCodeNotAccountFoundError ResponseCode = "EX002"
	ResponseCodeNotEnoughFundsError  ResponseCode = "EX007"
)
