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

func (t *Transcation) CalcTxID() string {
	hasher := sha256.New()
	for _, txIn := range t.TxIns {
		hasher.Write([]byte(txIn.TxOutId))
		hasher.Write([]byte(strconv.Itoa(txIn.TxOutIndex)))
	}

	for _, txOut := range t.TxOut {
		txOutAddr, _ := secp256k1.ParsePubKey(txOut.Address)
		hasher.Write(txOutAddr.SerializeCompressed())
		hasher.Write([]byte(strconv.Itoa(txOut.Amount)))
	}
	t.Id = hex.EncodeToString(hasher.Sum(nil))
	return t.Id

}

func GetCoinbaseTX(Amount int, Address *secp256k1.PublicKey, Height int) *Transcation {
	tx := &Transcation{}
	txIn := &TxIn{TxOutIndex: Height}

	txOut := &TxOut{
		Address: Address.SerializeCompressed(),
		Amount:  Amount,
	}

	tx.TxOut = append(tx.TxOut, txOut)
	tx.TxIns = append(tx.TxIns, txIn)
	tx.CalcTxID()
	return tx

}

func CheckUXTOandCheckSign(tx *Transcation, UXTOEntries *cache.Cache) (bool, error) {
	for _, txIn := range tx.TxIns {
		if uxto, found := UXTOEntries.Get(txIn.TxOutId + "-" + strconv.Itoa(txIn.TxOutIndex)); found {
			pubkeystr := uxto.(string)
			pubkeybyte, _ := hex.DecodeString(strings.Split(pubkeystr, "-")[0])
			pubkey, _ := secp256k1.ParsePubKey(pubkeybyte)
			sig, _ := ecdsa.ParseDERSignature(txIn.Sig)
			if sig.Verify([]byte(txIn.TxOutId), pubkey) == false {
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

func CheckAndSignTxIn(priv *secp256k1.PrivateKey, tx *Transcation, UXTOEntries *cache.Cache) error {
	for _, txIn := range tx.TxIns {
		if _, found := UXTOEntries.Get(txIn.TxOutId + "-" + strconv.Itoa(txIn.TxOutIndex)); found {
			sig := ecdsa.Sign(priv, []byte(txIn.TxOutId))
			txIn.Sig = sig.Serialize()
		} else {
			return errors.New("Can't Find Such UXTO!")
		}
	}
	return nil
}

func GenerateUXTO(FromAddr string, ToAddr string, Amount int, UXTOEntries *cache.Cache) ([]*TxIn, []*TxOut, error) {
	x := UXTOEntries.Items()

	balance := make(map[string]int)
	record := make(map[string][]string)
	for i, j := range x {
		str := j.Object.(string)
		acct := strings.Split(str, "-")
		temp, _ := strconv.Atoi(acct[1])
		balance[acct[0]] += temp
		record[acct[0]] = append(record[acct[0]], i+"-"+acct[1])
	}

	if balance[FromAddr] < Amount {
		return nil, nil, errors.New("not Enough Balance")
	} else {
		txIns := []*TxIn{}
		txOuts := []*TxOut{}
		changes := 0
		totalIn := 0
		toUseTX := make(map[string]int)
		for _, j := range record[FromAddr] {
			if totalIn < Amount {
				// j is the "TX"-"Idx"-"Amount"
				arr := strings.Split(j, "-")
				tempIn, _ := strconv.Atoi(arr[2])
				tempIdx, _ := strconv.Atoi(arr[1])
				totalIn += tempIn
				toUseTX[arr[0]] = tempIdx
			} else {
				break
			}
		}
		changes = totalIn - Amount

		for i, j := range toUseTX {
			txIn := &TxIn{
				TxOutId:    i,
				TxOutIndex: j,
				Sig:        nil,
			}
			txIns = append(txIns, txIn)
		}

		FromAddrbyte, _ := hex.DecodeString(FromAddr)
		FromAddrKey, _ := secp256k1.ParsePubKey(FromAddrbyte)
		ToAddrbyte, _ := hex.DecodeString(ToAddr)
		ToAddrKey, _ := secp256k1.ParsePubKey(ToAddrbyte)
		if changes > 0 {
			txOut := &TxOut{
				Address: FromAddrKey.SerializeCompressed(),
				Amount:  changes,
			}
			txOuts = append(txOuts, txOut)
		}
		txOut := &TxOut{
			Address: ToAddrKey.SerializeCompressed(),
			Amount:  Amount,
		}
		txOuts = append(txOuts, txOut)

		return txIns, txOuts, nil
	}

}
