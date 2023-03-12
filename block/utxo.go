package block

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

type UTXO struct {
	BlockChain *BlockChain
}

var (
	utxoPrefix = []byte("utxo-")
	prefLength = len(utxoPrefix)
)
func (utxo UTXO) FindUTXO(pubkey []byte) []TxtOutput {
	var utxos []TxtOutput
	db := utxo.BlockChain.Database
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next(){
             item := it.Item()
			 key,err := item.Value()
			 ErrorHandler(err)
			 out := DeSerializeoutput(key)
			 for _, outs := range out.Output {
				if !outs.IsLocked(pubkey) {
                     utxos = append(utxos, outs)
				}
			 }
		}
		return nil
	})
	ErrorHandler(err)
	
	return utxos
}

func (u UTXO) ReIndex() {
	db := u.BlockChain.Database
	u.DeletePrefix(utxoPrefix)
	UTXOset := u.BlockChain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error { //read and write only
		for UID, out := range UTXOset {
			key, err := hex.DecodeString(UID)
			ErrorHandler(err)
			key = append(utxoPrefix, key...)
			err = txn.Set(key, out.Serialize())
			ErrorHandler(err)
		}
		return nil
	})
	ErrorHandler(err)
}

func (utxo UTXO) Update(block *Block) {
	db := utxo.BlockChain.Database
	err := db.Update(func(txn *badger.Txn) error { //read and write only
		for _, tx := range block.Transaction {
			if tx.IsCoinBase() {   //bool
				for _, in := range tx.Input {
					outUpdates := TxtOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					ErrorHandler(err)
					v, err := item.Value()
					ErrorHandler(err)
					outs := DeSerializeoutput(v)
					for outIDx, out := range outs.Output {
						if outIDx != in.Out { //check if output equal to input
							outUpdates.Output = append(outUpdates.Output, out)
						}
					}
					if len(outUpdates.Output) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Set(inID, outUpdates.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}
			outupdate := TxtOutputs{}
			for _, in := range tx.Output {
				outupdate.Output = append(outupdate.Output, in)
			}
			inID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(inID, outupdate.Serialize()); err != nil {
				log.Panic(err)
			}

		}
		return nil
	})
	ErrorHandler(err)
}
func (utxo UTXO) CounterSet() int {
	db := utxo.BlockChain.Database
	counter := 0
	err := db.View(func(txn *badger.Txn) error {  //read only
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}
		return nil
	})
	ErrorHandler(err)
	return counter
}
func (utxo UTXO) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := utxo.BlockChain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			ErrorHandler(err)
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeSerializeoutput(v)
			for outIdx, out := range outs.Output {
				if out.IsLocked(pubKeyHash) && accumulated < amount {  //bool
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}

		}
		return nil
	})
	ErrorHandler(err)
	return accumulated, unspentOuts
}
func (utxo UTXO) FindUnspentTransactions(pubKeyHash []byte) []TxtOutput {
	var utxos []TxtOutput

	db := utxo.BlockChain.Database

	err := db.View(func(txn *badger.Txn) error { //read only
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			key, err := item.Value()
			ErrorHandler(err)
			out := DeSerializeoutput(key)
			for _, outs := range out.Output {
				if outs.IsLocked(pubKeyHash) {  //bool
					utxos = append(utxos, outs)
				}
			}
		}
		return nil
	})
	ErrorHandler(err)

	return utxos
}

func (u UTXO) DeletePrefix(prefix []byte) {
	deleteKey := func(keyDelete [][]byte) error {
		if err := u.BlockChain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keyDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}

			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectionSize := 100000
	u.BlockChain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keyDelete := make([][]byte, 0, collectionSize)
		collectedKey := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keyDelete = append(keyDelete, key)
			collectedKey++
			if collectedKey == collectionSize {
				if err := deleteKey(keyDelete); err != nil {
					log.Panic(err)
				}
				keyDelete = make([][]byte, 0, collectionSize)
				collectedKey = 0

			}
		}
		if collectedKey > 0 {
			if err := deleteKey(keyDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
