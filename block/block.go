package block

import (
	"bytes"
	"time"

	"encoding/gob"

	"log"
)

type Block struct {
	Transaction []*Transaction
	Hash        []byte
	PrevHash    []byte
	Ounce       int
	Timestamp   int64
	Height      int
}

func (b *Block) HashTransaction() []byte {
	var hashes [][]byte

	for _, tx := range b.Transaction {
		hashes = append(hashes, tx.Serialize())

	}
	tree := NewMerkleTree(hashes)
	return tree.RootNode.Data
}
func CreateBlock(tx []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{tx, []byte{}, prevHash, 0,time.Now().Unix(), height}
	pow := NewProof(block)
	Nounce, hash := pow.Run()
	block.Ounce = Nounce
	block.Hash = hash[:]
	return block
}

func Genesis(coinbase *Transaction) *Block {
	block := CreateBlock([]*Transaction{coinbase}, []byte{},0)
	return block
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer

	encode := gob.NewEncoder(&res)
	err := encode.Encode(b)
	ErrorHandler(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&block)
	ErrorHandler(err)
	return &block
}

func ErrorHandler(err error) {
	if err != nil {
		log.Panic(err)
	}
}
