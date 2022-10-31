package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/obasootom/Blockchain/block"
)


type Commandline struct{}

func (pri *Commandline) PrintLine() {

	fmt.Println("Usage:")
	fmt.Println("print - prints the block in the chain")
	fmt.Println("getbalance  -address ADDRESS - get the balance for address")
	fmt.Println("create blockchain -address ADDRESS create blockchain")
	fmt.Println("send -from FROM -to TO  -amount AMOUNT - Send amount")

}
func (pri *Commandline) validAgr() {
	if len(os.Args) < 2 {
		pri.PrintLine()
		runtime.Goexit()
	}
}
func (cli *Commandline) createblockchain(address string) {
	block := block.InitBlockchain(address)
	defer block.Database.Close()
	fmt.Println("Finished")
}
func (pri *Commandline) printChain() {
	chain := block.ContnueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		blocks := iter.Next()
		fmt.Printf("Previous hash : %x \n", blocks.PrevHash)
		fmt.Printf("Hash: %x \n", blocks.Hash)
		pow := block.NewProof(blocks)
		fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(blocks.PrevHash) == 0 {
			break
		}
	}
}
func (cli *Commandline) getBlockchain(address string) {
	chain := block.ContnueBlockchain(address)
	defer chain.Database.Close()
	balance := 0
	findtx := chain.FindTx(address)
	for _, out := range findtx {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}
func (cli *Commandline) send(from, to string, amount int) {
	chain := block.ContnueBlockchain(from)
	defer chain.Database.Close()
	tx := block.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*block.Transaction{tx})
	fmt.Println("Success")
}
func (pri *Commandline) Run() {
	pri.validAgr()

	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	getbalanceAddress := getbalanceCmd.String("address", "", "the address")
	createblockchainaddress := createchainCmd.String("address", "", "the address")
	sendTo := sendCmd.String("to", "", "destinaton wallet address")
	sendFrom := sendCmd.String("from", "", "source wallet")
	sendAmmount := sendCmd.Int("ammount", 0, "ammount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "createblockchain":
		err := createchainCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	default:
		pri.PrintLine()
		runtime.Goexit()
	}
	if getbalanceCmd.Parsed() {
		if *getbalanceAddress == "" { //
			getbalanceCmd.Usage()
			runtime.Goexit()
		}
		pri.getBlockchain(*getbalanceAddress)
	}
	if printChainCmd.Parsed() {
		pri.printChain()
	}
	if createchainCmd.Parsed() {
		if *createblockchainaddress == "" {
			createchainCmd.Usage()
			runtime.Goexit()
		}
		pri.createblockchain(*createblockchainaddress)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" && *sendTo == "" && *sendAmmount == 0 {
			runtime.Goexit()
		}
		pri.send(*sendFrom, *sendTo, *sendAmmount)

	}
}