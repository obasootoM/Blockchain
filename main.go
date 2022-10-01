package main

import (
	"fmt"

	"github.com/obasootom/Blockchain/block"
)

func main() {
	chain := block.InitBlockchain()

	chain.AddBlock("First block after genesis")
	chain.AddBlock("Second block after genesis")
	chain.AddBlock("Third block after genesis")

	for _, blocks := range chain.Blocks {
		fmt.Printf("Previous hash : %x \n", blocks.PrevHash)
		fmt.Printf("Data: %s \n", blocks.Data)
		fmt.Printf("Hash: %x \n", blocks.Hash)
	}
}
