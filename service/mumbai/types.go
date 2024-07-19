package mumbai

import (
	"errors"
)

var ErrInsufficientMumbaiWalletBalance = errors.New("insufficient mumbai wallet balance")
var ErrInsufficientUserWalletBalance = errors.New("insufficient user wallet balance")
var ErrGetBalance = errors.New("Mumbai get balance error")
var ErrGameCodeMapping = errors.New("game code mapping error")
