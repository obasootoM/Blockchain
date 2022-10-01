package block

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Data     []byte
	Hash     []byte
	PrevHash []byte
}

type Blockchain struct{
	Blocks []*Block //chains of blocks

}

func (block *Block) DeriveHash() {
	info := bytes.Join([][]byte{block.Data,block.PrevHash},[]byte{}) //contatinating an empty byte with bytes of block
	hash := sha256.Sum256(info)
	block.Hash = hash[:]
}

func  CreateBlock(data string, prevHash []byte) *Block{
	block := &Block{[]byte(data),[]byte{},prevHash}
	block.DeriveHash()
	return block
}

func (chain *Blockchain) AddBlock(data string) {
    prevBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
    chain.Blocks = append(chain.Blocks, new)
}

func Genesis() *Block{
	return CreateBlock("genesis",[]byte{})
}

func InitBlockchain() *Blockchain {
	return &Blockchain{[]*Block{Genesis()}}
}