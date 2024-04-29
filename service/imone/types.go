package imone

import "blgit.rfdev.tech/taya/game-service/imone"

var tayaGameCodeToImOneWalletCodeMapping = map[string]uint16{
	"impt":   imone.WalletCodePlayTech,
	"imslot": imone.WalletCodeIMSlot,
}
