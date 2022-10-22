package block

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./temp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
	//Blocks []*Block
}
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
		lastHash, err = item.Value()
		return err
	})
	ErrorHandler(err)
	newBlock := CreateBlock(data, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		ErrorHandler(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	ErrorHandler(err)
}

func InitBlockchain() *Blockchain {
	var lastHash []byte

	opts := badger.DefaultOptions

	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	ErrorHandler(err)
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("no Existing blockchain")
			genesis := Genesis()
			fmt.Println("genesis is proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			ErrorHandler(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			ErrorHandler(err)
			lastHash, err = item.Value()
			ErrorHandler(err)
		}
		return err
	})
	block := Blockchain{
		LastHash: lastHash,
		Database: db,
	}
	return &block
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{
		chain.LastHash,
		chain.Database,
	}
	return iter
}
func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		ErrorHandler(err)
		decode, err := item.Value()
		block = Deserialize(decode)
		return err
	})
	ErrorHandler(err)
	iter.CurrentHash = block.PrevHash
	return block
}
