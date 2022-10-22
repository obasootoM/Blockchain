package block

import (
	"bytes"
	"encoding/gob"

	"log"
)

type Block struct {
	Data     []byte
	Hash     []byte
	PrevHash []byte
	Ounce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte(data), []byte{}, prevHash, 0}
	pow := NewProof(block)
	Nounce, hash := pow.Run()
	block.Ounce = Nounce
	block.Hash = hash[:]
	return block
}

func Genesis() *Block {
	block := CreateBlock("genesis",[]byte{})
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
