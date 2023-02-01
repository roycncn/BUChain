package blockchain

type Block struct {
	index     int64
	hash      string
	prevHash  string
	timestamp int64
	data      string
}
