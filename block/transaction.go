package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/obasootom/Blockchain/wallet"
)

type Transaction struct {
	ID     []byte
	Input  []TxtInput
	Output []TxtOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}
func (tx Transaction) String() string {
	var line []string
	line = append(line, fmt.Sprintf("-- Transaction %x", tx.ID))
	for in, input := range tx.Input {
		line = append(line, fmt.Sprintf("input %d", in))
		line = append(line, fmt.Sprintf("txid %x", input.ID))
		line = append(line, fmt.Sprintf("out %d", input.Out))
		line = append(line, fmt.Sprintf("signature %x", input.Sig))
		line = append(line, fmt.Sprintf("pubkey %x", input.Pubkey))
	}
	for out, output := range tx.Output {
		line = append(line, fmt.Sprintf("output %x", out))
		line = append(line, fmt.Sprintf("value %d", output.Value))
		line = append(line, fmt.Sprintf("script %x", output.PubKey))
	}
	return strings.Join(line,"\n")
}
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]

}
func (tx Transaction) Sign(privateKey ecdsa.PrivateKey, prevTx map[string]Transaction) {
	if tx.IsCoinBase() {
		return
	}

	for _, in := range tx.Input {
		if prevTx[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("prev transaction does not exist")
		}
	}
	txCopy := tx.TrimCopy()

	for inID, in := range txCopy.Input {
		prevtx := prevTx[hex.EncodeToString(in.ID)]
		txCopy.Input[inID].Sig = nil
		txCopy.Input[inID].Pubkey = prevtx.Output[in.Out].PubKey
		txCopy.ID = tx.Hash()
		txCopy.Input[inID].Pubkey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		ErrorHandler(err)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Input[inID].Sig = signature
	}
}
func (tx *Transaction) TrimCopy() Transaction {
	var txinput []TxtInput
	var txoutput []TxtOutput
	for _, in := range tx.Input {
		txinput = append(txinput, TxtInput{in.ID, in.Out, nil, nil})

	}

	for _, out := range tx.Output {
		txoutput = append(txoutput, TxtOutput{out.Value, out.PubKey})
	}
	txCopy := Transaction{tx.ID, txinput, txoutput}
	return txCopy

}
func Coinbase(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("coin to %s", to)
	}
	txinp := TxtInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTx(100, to)

	tx := Transaction{nil, []TxtInput{txinp}, []TxtOutput{*txout}}
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
	walet, err := wallet.CreateWallet()
	ErrorHandler(err)
	w := walet.GetWallet(from)
	pubkeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOut := block.FindSpendableOutput(pubkeyHash, ammount)
	if acc < ammount {
		log.Panic("not enough fund")
	}
	for txid, outs := range validOut {
		txID, err := hex.DecodeString(txid)
		ErrorHandler(err)
		for _, out := range outs {
			inputs := TxtInput{txID, out, nil, w.PublicKey}
			input = append(input, inputs)

		}
	}
	output = append(output, *NewTx(ammount, to))
	if acc > ammount {
		output = append(output, *NewTx(acc-ammount, from))
	}
	tx := Transaction{
		nil,
		input,
		output,
	}
	tx.ID = tx.Hash()
	block.SignTransaction(&tx, w.PrivateKey)
	return &tx
}

func (tx *Transaction) Verify(prevTx map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}
	for _, in := range tx.Input {
		if prevTx[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("previous transacton does not exist")
		}
	}
	txCopy := tx.TrimCopy()
	curve := elliptic.P256()

	for inID, in := range tx.Input {
		prevTx := prevTx[hex.EncodeToString(in.ID)]
		txCopy.Input[inID].Sig = nil
		txCopy.Input[inID].Pubkey = prevTx.Output[in.Out].PubKey
		txCopy.ID = txCopy.Hash()
		txCopy.Input[inID].Pubkey = nil

		s := big.Int{}
		r := big.Int{}
		siglen := len(in.Sig)
		s.SetBytes(in.Sig[(siglen / 2):])
		r.SetBytes(in.Sig[:(siglen / 2)])

		x := big.Int{}
		y := big.Int{}

		pubkeylen := len(in.Pubkey)
		x.SetBytes(in.Pubkey[:(pubkeylen / 2)])
		y.SetBytes(in.Pubkey[(pubkeylen / 2):])

		pubRaw := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&pubRaw, txCopy.ID, &x, &y) == false {
			return false
		}
	}
	return true
}
