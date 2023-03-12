package block

import (
	"bytes"
	"encoding/gob"

	"github.com/obasootom/Blockchain/wallet"
)

type TxtOutput struct {
	Value      int
	PubKeyHash []byte
}
type TxtInput struct {
	ID     []byte
	Out    int
	Sig    []byte
	Pubkey []byte
}
type TxtOutputs struct {
	Output []TxtOutput
}

func (tx TxtOutputs) Serialize() []byte {
	var buffer bytes.Buffer

	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(tx)
	ErrorHandler(err)
	return buffer.Bytes()
}
func DeSerializeoutput(pubKey []byte) TxtOutputs {
	var out TxtOutputs
	decode := gob.NewDecoder(bytes.NewReader(pubKey))
	err := decode.Decode(&out)
	ErrorHandler(err)
	return out
}
func (t *Transaction) IsCoinBase() bool {
	return len(t.Input) == 1 && len(t.Input[0].ID) == 0 && t.Input[0].Out == -1
}

func (tx *TxtInput) UsesKey(pubkey []byte) bool {
	lockingHash := wallet.PublicKeyHash(tx.Pubkey)
	return bytes.Compare(pubkey, lockingHash) == 0

}

func (tx *TxtOutput) Lock(address []byte) {
	pubkeyHash := wallet.Base58Decode(address)
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-4]
	tx.PubKeyHash = pubkeyHash
}

func (tx *TxtOutput) IsLocked(pubkey []byte) bool {
	return bytes.Compare(tx.PubKeyHash, pubkey) == 0
}

func NewTx(value int, address string) *TxtOutput {
	output := &TxtOutput{value, nil}
	output.Lock([]byte(address))
	return output
}
