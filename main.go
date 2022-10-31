package main

import (
	"os"

	"github.com/obasootom/Blockchain/wallet"
	// "github.com/obasootom/Blockchain/cli"
)


func main() {
	defer os.Exit(0)
	// cl := cli.Commandline{}
	// cl.Run()
	wal :=  wallet.MakeWallet()
	wal.Address()

}
