package imone

import (
	"errors"

	"blgit.rfdev.tech/taya/game-service/imone"
)

var ErrTransferNegativeBalance = errors.New("imone not allowed to transfer negative sum")
var ErrGetBalance = errors.New("ImOne get balance error")
var ErrInsufficientImoneWalletBalance = errors.New("insufficient imone wallet balance")

var ErrGameCodeMapping = errors.New("game code mapping error")

var tayaGameCodeToImOneWalletCodeMapping = map[string]uint16{
	"impt":   imone.WalletCodePlayTech,
	"imslot": imone.WalletCodeIMSlot,
}
