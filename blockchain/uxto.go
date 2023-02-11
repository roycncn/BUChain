package blockchain

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

type UXTOEntry struct {
	Address    secp256k1.PublicKey
	Amount     int
	TxOutId    string
	TxOutIndex int
}
