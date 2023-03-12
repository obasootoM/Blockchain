package block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./temp/blocks_%s"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func retry(dir string, originalopt badger.Options) (*badger.DB, error) {
	path := filepath.Join(dir, "LOCK")
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf(`Removing:"LOCK"`, err)
	}
	retryopt := originalopt
	retryopt.Truncate = true
	db, err := badger.Open(retryopt)
	ErrorHandler(err)
	return db, err
}
func openDB(dir string, option badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(option); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, option); err == nil {
				fmt.Println("database unlocked, value log trucated")
				return db, nil
			}
			fmt.Println("could not unlock database")
		}
		return nil, err
	} else {
		return db, nil
	}
}
func ContinueBlockChain(address string) *BlockChain {
	path := fmt.Sprintf(dbPath, address)
	if !DBexists(path) { //bool
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	ErrorHandler(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
		lastHash, err = item.Value()

		return err
	})
	ErrorHandler(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

func InitBlockChain(address, nodeID string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeID)
	var lastHash []byte

	if !DBexists(path) { //bool
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	ErrorHandler(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := Coinbase(address, "genesis block")
		genesis := Genesis(cbtx)

		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		ErrorHandler(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	ErrorHandler(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(block *Block)  {

	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil{
			return nil
		}
		blockData := block.Serialize()
        err := txn.Set(block.Hash, blockData)
		ErrorHandler(err)
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
		lastHash, _ := item.Value()
		item, err = txn.Get(lastHash)
		ErrorHandler(err)
		lastBlockData,_ := item.Value()
		lastblock := Deserialize(lastBlockData)
		if block.Height > lastblock.Height  {
			err := txn.Set([]byte("lh"), block.Hash)
			ErrorHandler(err)
			chain.LastHash = block.Hash
		}
		return nil
	})
	ErrorHandler(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		ErrorHandler(err)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)

		return err
	})
	ErrorHandler(err)

	iter.CurrentHash = block.PrevHash

	return block
}

func (chain *BlockChain) FindUTXO() map[string]TxtOutputs {
	UTXOs := make(map[string]TxtOutputs)
	spentTxo := make(map[string][]int)
	iter := chain.Iterator()
	for {
		block := iter.Next()
		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Output:
			for outid, out := range tx.Output {
				if spentTxo[txID] != nil {
					for _, spentout := range spentTxo[txID] {
						if spentout == outid {
							continue Output
						}
					}
				}
				outs := UTXOs[txID]
				outs.Output = append(outs.Output, out)
				UTXOs[txID] = outs

			}
			if !tx.IsCoinBase() { //bool
				for _, in := range tx.Input {
					inTxID := hex.EncodeToString(in.ID)
					spentTxo[inTxID] = append(spentTxo[inTxID], in.Out)
				}
			}

		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXOs

}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transaction {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Input {
		prevTX, err := bc.FindTransaction(in.ID)
		ErrorHandler(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinBase() { //bool
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Input {
		prevTX, err := bc.FindTransaction(in.ID)
		ErrorHandler(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (chain *BlockChain) BestHeight() int {
      var lastBlock Block

	  err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
        lasthash, _ := item.Value()

		item, err = txn.Get(lasthash)
		ErrorHandler(err)
		lastHashBlock, _ := item.Value()
		lastBlock = *Deserialize(lastHashBlock)

		return nil
	  })
	  ErrorHandler(err)

	return lastBlock.Height
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
  var blocks [][]byte
  iter := chain.Iterator()

  for {
	block := iter.Next()

	if len(block.PrevHash)== 0 {
         break
	}
  }
  return blocks
}

func (chain *BlockChain) GetBlock(blockHash []byte) (*Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("block is not found")
		} else {
			blockData, _ := item.Value()
			block = *Deserialize(blockData)
		}
		return nil
	})
	ErrorHandler(err)
	return &block, nil
}

func (chain *BlockChain) MineBlock(transaction []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transaction {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("invalid transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
		lastBlockData, _ := item.Value()
		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		return err
	})
	ErrorHandler(err)

	newBlock := CreateBlock(transaction, lastHash, lastHeight+1)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())

		return err
	})

	return newBlock

}
