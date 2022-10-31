package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
)


type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

func NewPairKey() (ecdsa.PrivateKey,[]byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve,rand.Reader)
    if err != nil {
       log.Panic(err)
	}
	pub := append(private.X.Bytes(),private.Y.Bytes()...)
	return *private,pub

}