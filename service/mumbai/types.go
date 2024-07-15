package imone

import (
	"errors"
)

var ErrTransferNegativeBalance = errors.New("imone not allowed to transfer negative sum")
var ErrGetBalance = errors.New("Mumbai get balance error")
var ErrInsufficientImoneWalletBalance = errors.New("insufficient imone wallet balance")

var ErrGameCodeMapping = errors.New("game code mapping error")
