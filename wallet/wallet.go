package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	
	"log"

	"golang.org/x/crypto/ripemd160"
)
const(
	checksumLength = 4
	version = byte(0x00) //hexadecimal representation of zero
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

func MakeWallet() *Wallet {
	private,public := NewPairKey()
	wallet := Wallet{private,public}

	return &wallet
}

func PublicKeyHash(publickey []byte) []byte {
	pubHash := sha256.Sum256(publickey)
	hash := ripemd160.New()
	_, err := hash.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}
   pubHahpri := hash.Sum(nil)
   return pubHahpri
}


func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHah := sha256.Sum256(firstHash[:])

	return secondHah[:checksumLength]
}

func (w *Wallet) Address() []byte {
	publicHash := PublicKeyHash(w.PublicKey)
	versionHash := append([]byte{version},publicHash...)
	checksumHash := Checksum(versionHash)
	fullHash := append(versionHash, checksumHash...)
	address  := Base58Ecode(fullHash)

	// fmt.Printf("pub key %x\n",w.PublicKey)
	// fmt.Printf("pub hash %x\n",publicHash)
	// fmt.Printf("address %x\n", address)
	return address
}