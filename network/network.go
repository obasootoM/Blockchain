package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/obasootom/Blockchain/block"
	"github.com/vrecan/death/v3"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	minerAddress    string
	knownNodes      = []string{"localhost:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]block.Transaction)
)

type Addres struct {
	AddFrom []string
}

type Block struct {
	AddFrom string
	Block   []byte
}

type GetBlock struct {
	AddFrom string
}

type GetData struct {
	AddFrom string
	Type    string
	ID      []byte
}

type Inv struct {
	AddFrom string
	Type    string
	Item    [][]byte
}

type Tx struct {
	AddFrom     string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddFrom    string
}

func CmdToByte(cmd string) []byte {
	var bytes [commandLength]byte
	for i, v := range cmd {
		bytes[i] = byte(v)
	}
	return bytes[:]
}

func ByteToCmd(cmd []byte) string {
	var bytes []byte

	for _, i := range cmd {
		if i != 0x0 {
			bytes = append(bytes, i)
		}
	}
	return fmt.Sprintf("%s", bytes)
}

func GonEncoder(data interface{}) []byte {
	var buffer bytes.Buffer

	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

func HandleConnection(conn net.Conn, chain *block.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := ByteToCmd(req[:commandLength])
	fmt.Printf("Recieved %s command \n", command)

	switch command {
	case "addr":
		HandleAddress(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "version":
		HandleVersion(req, *chain)
	case "tx":
		HandleTx(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "getblock":
		HandleGetBlock(req, chain)
	default:
		fmt.Println("unknown command")
	}
}
func StartServer(nodeID, mineAdrress string) {
	nodeAddress = fmt.Sprintf("localhost: %s", nodeID)
	mineAdrress = minerAddress
	listen, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listen.Close()
	chain := block.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAddress != knownNodes[0] {
		SendVersion(knownNodes[0], chain)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)
	}
}
func SendData(add string, data []byte) {
	conn, err := net.Dial(protocol, add)
	if err != nil {
		fmt.Printf("%s add unavailable \n", add)
		var updateNode []string

		for _, node := range knownNodes {
			if node != add {
				updateNode = append(updateNode, node)
			}
		}
		knownNodes = updateNode
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func SendAdd(add string) {
	nodes := Addres{knownNodes}
	nodes.AddFrom = append(nodes.AddFrom, nodeAddress)
	payload := GonEncoder(nodes)
	request := append(CmdToByte("addres"), payload...)
	SendData(add, request)

}

func SendBlock(add string, chain *block.Block) {
	data := Block{nodeAddress, chain.Serialize()}
	payload := GonEncoder(data)
	request := append(CmdToByte("block"), payload...)
	SendData(add, request)
}

func SendInv(add, kind string, item [][]byte) {
	inventory := Inv{nodeAddress, kind, item}
	payload := GonEncoder(inventory)
	request := append(CmdToByte("inventory"), payload...)
	SendData(add, request)
}

func SendTx(add string, tx *block.Transaction) {
	data := Tx{add, tx.Serialize()}
	payload := GonEncoder(data)
	request := append(CmdToByte("tx"), payload...)
	SendData(add, request)
}

func SendVersion(add string, chain *block.BlockChain) {
	bestHeight := chain.BestHeight()
	payload := GonEncoder(Version{version, bestHeight, nodeAddress})
	request := append(CmdToByte("version"), payload...)
	SendData(add, request)
}

func SendGetData(add, kind string, id []byte) {
	payload := GonEncoder(GetData{nodeAddress, kind, id})
	request := append(CmdToByte("getdata"), payload...)
	SendData(add, request)
}

func SendGetBlock(add string) {
	payload := GonEncoder(GetBlock{nodeAddress})
	request := append(CmdToByte("getblock"), payload...)
	SendData(add, request)
}

func HandleAddress(request []byte) {
	var buff bytes.Buffer
	var address Addres
	buff.Write(request[:commandLength])
	encode := gob.NewDecoder(&buff)
	err := encode.Decode(&address)
	if err != nil {
		log.Panic(err)
	}
	knownNodes = append(knownNodes, address.AddFrom...)
	fmt.Printf("there are %d of nodes", len(knownNodes))
	RequestBlock()
}

func RequestBlock() {
	for _, node := range knownNodes {
		SendGetBlock(node)
	}
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

func HandleBlock(request []byte, chain *block.BlockChain) {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[:commandLength])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	blocks := block.Deserialize(blockData)
	fmt.Println("Recived a new block")
	chain.AddBlock(blocks)
	fmt.Printf("Add block %x\n", blocks.Hash)
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		utxo := block.UTXO{chain}
		utxo.ReIndex()
	}
}

func HandleGetBlock(request []byte, chain *block.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlock

	buff.Write(request[:commandLength])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := chain.GetBlockHashes()
	SendInv(payload.AddFrom, "block", blocks)
}

func HandleGetData(request []byte, chain *block.BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		SendBlock(payload.AddFrom, block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]
		SendTx(payload.AddFrom, &tx)
	}
}

func HandleVersion(request []byte, chain block.BlockChain) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(payload)
	if err != nil {
		log.Panic(err)
	}

	bestHeight := chain.BestHeight()
	otherHeight := payload.BestHeight

	if bestHeight < otherHeight {
		SendGetBlock(payload.AddFrom)
	} else if bestHeight > otherHeight {
		SendVersion(payload.AddFrom, &chain)
	}
	if !NodeIsKnown(payload.AddFrom) {
		knownNodes = append(knownNodes, payload.AddFrom)
	}
}

func NodeIsKnown(add string) bool {
	for _, node := range knownNodes {
		if node == add {
			return true
		}
	}
	return false
}

func HandleTx(request []byte, chain *block.BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(payload)
	if err != nil {
		log.Panic(err)
	}
	txData := payload.Transaction
	tx := block.DeserailzationTransation(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	fmt.Printf("%s,%d\n", nodeAddress, len(memoryPool))

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node == nodeAddress && node != payload.AddFrom {
				SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(minerAddress) > 0 {
			MinerTx(chain)
		}
	}
}

func MinerTx(chain *block.BlockChain) {
	var txs []*block.Transaction

	for id := range memoryPool {
		fmt.Printf("tx: %s \n", memoryPool[id].ID)
		tx := memoryPool[id]
		if chain.VerifyTransaction(&tx) {
			txs = append(txs, &tx)
		}
	}
	if len(txs) == 0 {
		fmt.Println("All transation are invalid")
		return
	}
	cbtx := block.Coinbase(minerAddress, "")
	txs = append(txs, cbtx)

	newblock := chain.MineBlock(txs)
	utxoSet := block.UTXO{chain}
	utxoSet.ReIndex()

	fmt.Println("new block is mined")

	for _, txss := range txs {
		txid := hex.EncodeToString(txss.ID)
		delete(memoryPool, txid)
	}

	for _, node := range knownNodes {
		if node != nodeAddress {
			SendInv(node, "block", [][]byte{newblock.Hash})
		}
	}
	if len(memoryPool) > 0 {
		MinerTx(chain)
	}
}

func HandleInv(request []byte, chain *block.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	decode := gob.NewDecoder(&buff)
	err := decode.Decode(payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("rRecieved inventory with %d %s \n", len(payload.Item), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Item

		blockHash := payload.Item[0]
		SendGetData(payload.AddFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Item[0]
		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddFrom, "tx", txID)
		}
	}
}

func CloseDB(chain *block.BlockChain) {
	db := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	db.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
