package mumbai

import (
	"errors"
)

var ErrGetBalance = errors.New("Mumbai get balance error")
var ErrGameCodeMapping = errors.New("game code mapping error")

type ResponseCode string

const (
	ResponseCodeNotAccountFoundError ResponseCode = "EX002"
)
