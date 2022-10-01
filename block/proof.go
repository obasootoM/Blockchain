package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	pow := &ProofOfWork{b, target}
	return pow
}

func (p *ProofOfWork) InitData(ounce int) []byte {
	data := bytes.Join([][]byte{
		p.Block.PrevHash,
		p.Block.Data,
		ToDec(int64(ounce)),
		ToDec(int64(-Difficulty)),
	}, []byte{})

	return data
}

func ToDec(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func (p *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	ounce := 0

	for ounce < math.MaxInt64 {
		data := p.InitData(ounce)
		hash = sha256.Sum256(data)
        fmt.Printf("\r %x",hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(p.Target)== -1 {
			break
		}else {
			ounce ++
		}
	}
	fmt.Println()
	return ounce, hash[:]
}