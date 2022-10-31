package wallet

import (
	"log"

	"github.com/mr-tron/base58/base58"
)


func Base58Ecode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
        log.Fatal(err)
	}
	return decode
}