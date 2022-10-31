package main

import (
	"os"


	"github.com/obasootom/Blockchain/cli"
)


func main() {
	defer os.Exit(0)
	cl := cli.Commandline{}
	cl.Run()
}
