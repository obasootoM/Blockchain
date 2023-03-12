package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00) //hexadecimal representation of zero
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey //private key
	PublicKey  []byte           //public key
}

func NewPairKey() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pub := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pub

}

func MakeWallet() *Wallet {
	private, public := NewPairKey()
	wallet := Wallet{private, public}

	return &wallet
}

func PublicKeyHash(publickey []byte) []byte {
	pubHash := sha256.Sum256(publickey)
	hash := ripemd160.New()
	_, err := hash.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}
	pubHashkey := hash.Sum(nil)
	return pubHashkey
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func (w *Wallet) Address() []byte {
	publicHash := PublicKeyHash(w.PublicKey)
	versionHash := append([]byte{version}, publicHash...)
	checksumHash := Checksum(versionHash)
	fullHash := append(versionHash, checksumHash...)
	address := Base58Encode(fullHash)

	// fmt.Printf("pub key %x\n",w.PublicKey)
	// fmt.Printf("pub hash %x\n",publicHash)
	// fmt.Printf("address %x\n", address)
	return address
}

func Validate(address string) bool {
	pubkeyHash := Base58Decode([]byte(address))
	actualChecksum := pubkeyHash[len(pubkeyHash)-checksumLength:]
	versions := pubkeyHash[0]
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{versions}, pubkeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
