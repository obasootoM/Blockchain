package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletfile = "./temp/wallet.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func (w *Wallets) Savefile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encode := gob.NewEncoder(&content)
	err := encode.Encode(w)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletfile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func (w *Wallets) Loadfile() error {
	if _, err := os.Stat(walletfile); os.IsNotExist(err) {
		return err
	}
	var wallets Wallets
	filecontent, err := ioutil.ReadFile(walletfile)
	if err != nil {
		log.Panic(err)
	}
	gob.Register(elliptic.P256())
	decode := gob.NewDecoder(bytes.NewReader(filecontent))
	err = decode.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	w.Wallets = wallets.Wallets
	return nil
}

func CreateWallet() (*Wallets, error) {
	wallet := Wallets{}
	wallet.Wallets = map[string]*Wallet{}
	err := wallet.Loadfile()
	return &wallet, err
}

func (w *Wallets) GetWallet(address string) Wallet {
	return *w.Wallets[address]

}

func (w *Wallets) GetAllAddress() []string {
	var address []string
	for addressess := range w.Wallets {
		address = append(address, addressess)
	}
	return address
}
func (w *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	w.Wallets[address] = wallet
	return address
}
