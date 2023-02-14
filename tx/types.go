package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/patrickmn/go-cache"
	"strconv"
	"strings"
)

type Transcation struct {
	Id    string
	TxIns []*TxIn
	TxOut []*TxOut
}

type TxIn struct {
	TxOutId    string
	TxOutIndex int
	Sig        []byte
}

type TxOut struct {
	Address *secp256k1.PublicKey
	Amount  int
}

func (t *Transcation) CalcTxID() string {
	hasher := sha256.New()
	for _, txIn := range t.TxIns {
		hasher.Write([]byte(txIn.TxOutId))
		hasher.Write([]byte(strconv.Itoa(txIn.TxOutIndex)))
	}

	for _, txOut := range t.TxOut {
		hasher.Write(txOut.Address.SerializeCompressed())
		hasher.Write([]byte(strconv.Itoa(txOut.Amount)))
	}
	t.Id = hex.EncodeToString(hasher.Sum(nil))
	return t.Id

}

func GetCoinbaseTX(Amount int, Address *secp256k1.PublicKey, Height int) *Transcation {
	tx := &Transcation{}
	txIn := &TxIn{TxOutIndex: Height}

	txOut := &TxOut{
		Address: Address,
		Amount:  Amount,
	}

	tx.TxOut = append(tx.TxOut, txOut)
	tx.TxIns = append(tx.TxIns, txIn)
	tx.CalcTxID()
	return tx

}

func CheckUXTOandCheckSign(tx *Transcation, UXTOEntries *cache.Cache) (bool, error) {
	for _, txIn := range tx.TxIns {
		if uxto, found := UXTOEntries.Get(txIn.TxOutId + "-" + string(txIn.TxOutIndex)); found {
			pubkeystr := uxto.(string)
			pubkeybyte, _ := hex.DecodeString(strings.Split(pubkeystr, "-")[0])
			pubkey, _ := secp256k1.ParsePubKey(pubkeybyte)
			txidbyte, _ := hex.DecodeString(txIn.TxOutId)
			sig, _ := ecdsa.ParseDERSignature(txIn.Sig)
			if sig.Verify(txidbyte, pubkey) == false {
				return false, errors.New("Sig Wrong")
			}
		} else {
			return false, errors.New("Can't Find Such UXTO!")
		}
	}
	if tx.CalcTxID() != tx.Id {
		return false, errors.New("Tx ID wrong")
	}

	return true, nil
}

func CheckAndSignTxIn(priv *secp256k1.PrivateKey, tx *Transcation, txInIndex int, UXTOEntries *cache.Cache) (*ecdsa.Signature, error) {
	for _, txIn := range tx.TxIns {
		if _, found := UXTOEntries.Get(txIn.TxOutId + "-" + string(txIn.TxOutIndex)); found {

		} else {
			return &ecdsa.Signature{}, errors.New("Can't Find Such UXTO!")
		}
	}
	sig := ecdsa.Sign(priv, []byte(tx.Id))
	return sig, nil
}
