package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/obasootom/Blockchain/block"
	"github.com/obasootom/Blockchain/wallet"
)

type Commandline struct{}

func (pri *Commandline) PrintLine() {

	fmt.Println("Usage:")
	fmt.Println("print - prints the block in the chain")
	fmt.Println("getbalance  -address ADDRESS - get the balance for address")
	fmt.Println("create blockchain -address ADDRESS create blockchain")
	fmt.Println("send -from FROM -to TO  -amount AMOUNT - Send amount")
	fmt.Println("createwallet - create a new wallet")
	fmt.Println("listaddress - list the address in our wallet file")
	fmt.Println("Reindexutxo - Rebuilds the UTXO set")

}
func (pri *Commandline) validAgr() {
	if len(os.Args) < 2 {
		pri.PrintLine()
		runtime.Goexit()
	}
}
func (cli *Commandline) createblockchain(address string) {
	if !wallet.Validate(address) {
		log.Panic("address is  not valid")
	}
	blocks := block.InitBlockChain(address)
	defer blocks.Database.Close()
	utxo := block.UTXO{BlockChain: blocks}
	utxo.ReIndex()
	fmt.Println("Finished")
}
func (pri *Commandline) printChain() {
	chain := block.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		blocks := iter.Next()
		fmt.Printf("Hash : %x \n", blocks.Hash)
		fmt.Printf("prev.Hash: %x \n", blocks.PrevHash)
		pow := block.NewProof(blocks)
		fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range blocks.Transaction {
			fmt.Println(tx)
		}
		fmt.Println()
		if len(blocks.PrevHash) == 0 {
			break
		}
	}
}
func (cli *Commandline) createwallet() {
	wallet, _ := wallet.CreateWallet()
	address := wallet.AddWallet()
	wallet.Savefile()
	fmt.Printf("new address is %s\n", address)
}
func (cli *Commandline) listaddress() {
	wallet, _ := wallet.CreateWallet()
	address := wallet.GetAllAddress()
	for _, addresses := range address {
		fmt.Println(addresses)
	}
}

func (cli *Commandline) reindex() {
	chain := block.ContinueBlockChain("")
	defer chain.Database.Close()
	utxo := block.UTXO{BlockChain: chain}
	utxo.ReIndex()

	count := utxo.CounterSet()
	fmt.Printf("Done! there are %d transaction of utxo set\n", count)
}
func (cli *Commandline) getbalance(address string) {
	if !wallet.Validate(address) {
		log.Panic("address is not valid")
	}
	chain := block.ContinueBlockChain(address)
	utxo := block.UTXO{BlockChain: chain}
	defer chain.Database.Close()
	balance := 0
	pubkeyHash := wallet.Base58Encode([]byte(address))
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-4]

	findtx := utxo.FindUTXO(pubkeyHash)
	for _, out := range findtx {
		balance += out.Value
	}
	fmt.Printf("Balance of %s:  %d\n", address, balance)
}
func (cli *Commandline) send(from, to string, amount int) {
	if !wallet.Validate(to) {
		log.Panic("address is not valid")
	}
	if !wallet.Validate(from) {
		log.Panic("address is not valid")
	}
	chain := block.ContinueBlockChain(from)
	defer chain.Database.Close()
	utxo := block.UTXO{BlockChain: chain}
	tx := block.NewTransaction(from, to, amount, &utxo)
	cbtx := block.Coinbase(to, "")
	blocks := chain.AddBlock([]*block.Transaction{cbtx,tx})
	utxo.Update(blocks)
	fmt.Println("Success")
}
func (pri *Commandline) Run() {
	pri.validAgr()

	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	creatwewalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listaddresscmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	reindexcmd := flag.NewFlagSet("reindex", flag.ExitOnError)

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
	case "createwallet":
		err := creatwewalletcmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "listaddress":
		err := listaddresscmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	case "reindex":
		err := 	reindexcmd.Parse(os.Args[2:])
		block.ErrorHandler(err)
	default:
		pri.PrintLine()
		runtime.Goexit()
	}
	if reindexcmd.Parsed() {
		pri.reindex()
	}
	if getbalanceCmd.Parsed() {
		if *getbalanceAddress == "" { //
			getbalanceCmd.Usage()
			runtime.Goexit()
		}
		pri.getbalance(*getbalanceAddress)
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
	if creatwewalletcmd.Parsed() {
		pri.createwallet()
	}
	if listaddresscmd.Parsed() {
		pri.listaddress()
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmmount <= 0 {
			runtime.Goexit()
		}
		pri.send(*sendFrom, *sendTo, *sendAmmount)

	}
}
