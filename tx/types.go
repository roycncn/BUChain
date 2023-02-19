package tx

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
	Address []byte
	Amount  int
}
