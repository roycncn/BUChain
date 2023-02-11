package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/patrickmn/go-cache"
	"strconv"
)

type Transcation struct {
	Id    string
	TxIns []TxIn
	TxOut []TxOut
}

type TxIn struct {
	TxOutId    string
	TxOutIndex int
	Sig        []byte
}

type TxOut struct {
	Address secp256k1.PublicKey
	Amount  int
}

func (t *Transcation) calcTxID() string {
	hasher := sha256.New()
	for _, txIn := range t.TxIns {
		hasher.Write([]byte(txIn.TxOutId))
		hasher.Write([]byte(strconv.Itoa(txIn.TxOutIndex)))
	}

	for _, txOut := range t.TxOut {
		hasher.Write(txOut.Address.SerializeCompressed())
		hasher.Write([]byte(strconv.Itoa(txOut.Amount)))
	}
	result := hex.EncodeToString(hasher.Sum(nil))
	return result
}

func CheckAndSignTxIn(priv *secp256k1.PrivateKey, tx *Transcation, txInIndex int, UXTOEntries *cache.Cache) (*ecdsa.Signature, error) {
	for _, txIn := range tx.TxIns {
		if _, found := UXTOEntries.Get(txIn.TxOutId + "_" + string(txIn.TxOutIndex)); found {

		} else {
			return &ecdsa.Signature{}, errors.New("Can't Find Such UXTO!")
		}
	}
	sig := ecdsa.Sign(priv, []byte(tx.Id))
	return sig, nil
}
