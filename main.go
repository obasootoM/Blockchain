package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/obasootom/Blockchain/block"
)

type Commandline struct {
	blockChain *block.Blockchain
}

func (pri *Commandline) PrintLine() {

	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println("print - prints the block in the chain")

}
func (pri *Commandline) validAgr() {
	if len(os.Args) < 2 {
		pri.PrintLine()
		runtime.Goexit()
	}
}
func (pri *Commandline) addBlock(data string) {
	pri.blockChain.AddBlock(data)
	fmt.Println("block added")
}
func (pri *Commandline) printChain() {
	iter := pri.blockChain.Iterator()
	for {
		blocks := iter.Next()
		fmt.Printf("Previous hash : %x \n", blocks.PrevHash)
		fmt.Printf("Data: %s \n", blocks.Data)
		fmt.Printf("Hash: %x \n", blocks.Hash)
		pow := block.NewProof(blocks)
		fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(blocks.PrevHash) == 0 {
			break
		}
	}
}
func (pri *Commandline) run() {
	pri.validAgr()
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	default:
		pri.PrintLine()
		runtime.Goexit()
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		pri.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		pri.printChain()
	}
}
func main() {
	defer os.Exit(0)
	chain := block.InitBlockchain()
	defer chain.Database.Close()
	pri := Commandline{chain}
	pri.run()

}
