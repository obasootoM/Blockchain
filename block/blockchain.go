package block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./temp/blocks"
	dbFile      = "./temp/blocks/MANIFEST"
	genesisData = "first transaction from genesis"
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

func (chain Blockchain) FindTransaction(id []byte) (Transaction,error) {
	block := chain.Iterator()

	for {
		bloc := block.Next()

		for _, tx := range bloc.Transaction {
			if bytes.Compare(tx.ID,id) == 0 {
				return *tx,nil
			}
		}
		if len(bloc.PrevHash) == 0 {
			break
		}

	}
	return Transaction{},errors.New("transaction does not exit")
}

func (chain *Blockchain) AddBlock(tx []*Transaction) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error { //read only
		item, err := txn.Get([]byte("lh"))
		ErrorHandler(err)
		lastHash, err = item.Value()
		return err
	})
	ErrorHandler(err)
	newBlock := CreateBlock(tx, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error { //read and write
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		ErrorHandler(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	ErrorHandler(err)
}



func DBExist() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}



func InitBlockchain(address string) *Blockchain {
	var lastHash []byte
	if !DBExist() {
		fmt.Println("blockchain already exist")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions

	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	ErrorHandler(err)
	err = db.Update(func(txn *badger.Txn) error {
		ct := Coinbase(address, genesisData)
		genesis := Genesis(ct)
		fmt.Println("genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		ErrorHandler(err)
		err = txn.Set([]byte("lh"), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})
	block := Blockchain{
		LastHash: lastHash,
		Database: db,
	}
	return &block
}



func ContnueBlockchain(address string) *Blockchain {
	if !DBExist()  {
		fmt.Println("NO existing blockchain found, create one")
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
	chain := Blockchain{
		lastHash,
		db,
	}
	return &chain
}

func (bl *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)
	for _,in := range tx.Input {
		prevTx ,err := bl.FindTransaction(in.ID)
		ErrorHandler(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	tx.Sign(privateKey,prevTxs)
}


func (bl *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTx := make(map[string]Transaction)
	for _,in := range tx.Input{
		prevT, err := bl.FindTransaction(in.ID) 
		ErrorHandler(err)
		prevTx[hex.EncodeToString(prevT.ID)] = prevT
	}
	return tx.Verify(prevTx)
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


func (block *Blockchain) FindUnspentTransaction(pubKey []byte) []Transaction {
	var UspentTx []Transaction
	spentTx := make(map[string][]int)
	iter := block.Iterator()
	for {
		chain := iter.Next()
		for _, tx := range chain.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outID, out := range tx.Output {
				if spentTx[txID] != nil {
					for _, spentOut := range spentTx[txID] {
						if spentOut == outID {
							continue Outputs
						}

					}
				}
				if out.IsLocked(pubKey) {
					UspentTx = append(UspentTx, *tx)
				}
			}
			if !tx.IsCoinBase() {
				for _, in := range tx.Input {
					if in.UsesKey(pubKey) {
						inTx := hex.EncodeToString(in.ID)
						spentTx[inTx] = append(spentTx[inTx], in.Out)
					}
				}
			}
		}
		if len(chain.PrevHash) == 0 {
			break
		}
	}
	return UspentTx
}



func (block *Blockchain) FindTx(pubKey []byte) []TxtOutput {
	var txOutput []TxtOutput
	unspentransaction := block.FindUnspentTransaction(pubKey)
     for _,tx := range unspentransaction {
		for _, out := range tx.Output{
			if out.IsLocked(pubKey) {
				txOutput = append(txOutput, out)

			}
		}
	 }
	return txOutput
}

func (block *Blockchain) FindSpendableOutput(pubKey []byte, ammount int) (int, map[string][]int) {
	unspentOut := make(map[string][]int)
	unspentTx := block.FindUnspentTransaction(pubKey)
	accumulated := 0
	Work:
       for _,tx := range unspentTx {
		txID := hex.EncodeToString(tx.ID)
           for outID, out := range tx.Output {
			if out.IsLocked(pubKey) && accumulated < ammount {
				accumulated += out.Value
				unspentOut[txID] = append(unspentOut[txID],outID)

				if accumulated >= ammount {
					break Work
				}
			}
		   }
	   }

	return accumulated,unspentOut

}


