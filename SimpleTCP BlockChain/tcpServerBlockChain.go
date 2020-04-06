package main
import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Structure of the Block
type MyBlock struct {
	Index     int
	Timestamp string
	BlockNo   int
	Hash      string
	PrevHash  string
}

// Initialize a new Block chain(Type->MyBlock)
var Blockchain [] MyBlock

// blockServer to handel the upcoming blocks to the chain
var blockServer chan [] MyBlock  //Initialize My Block Type chan(Array) 
var mutex = &sync.Mutex{}



// Check if the Block is valid or not
func isValidBlock(newBlock, prevBlock MyBlock) bool {
	if prevBlock.Index+1 != newBlock.Index {
		return false
	}

	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

// method to replace the chain with new block
func replaceChain(newBlocks []MyBlock) {
	mutex.Lock()
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
	mutex.Unlock()
}

// method to calcualte the hash value (SHA-256)
func calculateHash(block MyBlock) string {
	record := string(block.Index) + block.Timestamp + string(block.BlockNo) + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func createNewBlock(prevBlock MyBlock, BlockNo int) (MyBlock, error) {

	var newBlock MyBlock

	time := time.Now()

	newBlock.Index = prevBlock.Index + 1
	newBlock.Timestamp = time.String()
	newBlock.BlockNo = BlockNo
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}


// Main Method
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	blockServer = make(chan [] MyBlock)  //Make the MyBlock Type Chain 

	// creat the initial block
	time := time.Now()
	genesisBlock := MyBlock{0, time.String(), 0, "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	tcpPort := os.Getenv("ADDR")

	// start TCP and serve TCP server
	tcpServer, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP  Server Listening on port :", tcpPort)
	defer tcpServer.Close()

	for {
		connection, err := tcpServer.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(connection)
	}

}

// Method to handle new block
func handleConnection(connection net.Conn) {

	defer connection.Close()

	io.WriteString(connection, "Enter a new Block No:") //Get new block no from the user

	scanner := bufio.NewScanner(connection)

	
	go func() {
		for scanner.Scan() {
			// Check if the user entered no is a integer or not
			blockNo, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v Not a number: %v", scanner.Text(), err)
				continue
			}
			// If it is a number create the new block according to the block no
			newBlock, err := createNewBlock(Blockchain[len(Blockchain)-1], blockNo)
			if err != nil {
				log.Println(err)
				continue
			}

			// Check is new block is valid one or not
			if isValidBlock(newBlock,Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			// Send block chain to block server
			blockServer <- Blockchain
			io.WriteString(connection, "\nEnter a new Block No:")
		}
	}()

	
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()

			// Print Json Output
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
			io.WriteString(connection, string(output))
		}
	}()

	// Blank Identifires to continue the app (Continue getting user inputs)
	for _ = range blockServer {
		spew.Dump(Blockchain)

	}

}
