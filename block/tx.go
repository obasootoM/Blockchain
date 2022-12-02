package block

import (
	"bytes"

	"github.com/obasootom/Blockchain/wallet"
)



type TxtOutput struct {
	Value  int
	PubKey []byte
}
type TxtInput struct {
	ID  []byte
	Out int
	Sig []byte
	Pubkey []byte
}

func (t *Transaction) IsCoinBase() bool {
	return len(t.Input) == 1 && len(t.Input[0].ID) == 0 && t.Input[0].Out == -1
}


func (tx TxtInput) UsesKey(pubkey []byte) bool {
	lockingHash := wallet.PublicKeyHash(tx.Pubkey)
	return bytes.Compare(pubkey,lockingHash) == 0

}

func (tx TxtOutput) Lock(address []byte) {
    pubkeyHash := wallet.Base58Decode(address)
	pubkeyHash = pubkeyHash[1:len(pubkeyHash) -4]
	tx.PubKey = pubkeyHash
}

func (tx TxtOutput) IsLocked(pubkey []byte) bool {
  return bytes.Compare(tx.PubKey, pubkey) == 0
}

func NewTx(value int, address string) *TxtOutput {
   output := &TxtOutput{Value: value, PubKey: nil}
  output.Lock([]byte(address))
  return output
}

