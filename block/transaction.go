package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID     []byte
	Input  []TxtInput
	Output []TxtOutput
}



func Coinbase(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("coin to %s", to)
	}
	txinp := TxtInput{[]byte{}, -1, data}
	txout := TxtOutput{100, to}

	tx := Transaction{nil, []TxtInput{txinp}, []TxtOutput{txout}}
	tx.SetId()
	return &tx

}

func (t *Transaction) SetId() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(t)
	ErrorHandler(err)
	hash = sha256.Sum256(encoded.Bytes())
	t.ID = hash[:]
}



func NewTransaction(from, to string, ammount int, block *Blockchain) *Transaction {
	var output []TxtOutput
	var input []TxtInput

	acc, validOut := block.FindSpendableOutput(from, ammount)
	if acc < ammount {
		log.Panic("not enough fund")
	}
	for txid, outs := range validOut {
		txID, err := hex.DecodeString(txid)
		ErrorHandler(err)
		for _, out := range outs {
			inputs := TxtInput{txID, out, from}
			input = append(input, inputs)

		}
	}
	output = append(output, TxtOutput{ammount, to})
	if acc > ammount {
		output = append(output, TxtOutput{acc - ammount, to})
	}
	tx := Transaction{
		nil,
		input,
		output,
	}
	tx.SetId()
	return &tx
}
