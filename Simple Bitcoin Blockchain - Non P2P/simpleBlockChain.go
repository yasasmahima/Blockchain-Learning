// Simple (Cryptocurrencies)Bitcoin Block Chain with SHA-256 Hashing
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Create the structure of a Block 
type MyBlock struct {
	Index      int
	Timestamp  string
	BlockNo    int
	Hash       string
	PrevHash   string
}

//Structure of the Post Message of a Block
type Message struct {
	BlockNo int
}

// Initialize a new Block chain(Type->MyBlock)
var Blockchain [] MyBlock

// Function for calculate the hash value of each block - Convert the each blocks values in to fixed length string  
func calculateHashValue(block MyBlock) string {
	record := string(block.Index) + block.Timestamp + string(block.BlockNo) + block.PrevHash
	hash := sha256.New()      //use SHA-256 Hashing 
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

// Function for create a new block
func generateBlock(prevBlock MyBlock,BlockNo int) (MyBlock, error) {

	// Initialize a block
	var newBlock MyBlock

	time := time.Now()  //Get Block created time

	newBlock.Index = prevBlock.Index + 1  //Index of the new block = Index of old block+1
	newBlock.Timestamp = time.String()
	newBlock.BlockNo = BlockNo            //Set Block No 
	newBlock.PrevHash = prevBlock.Hash    //PrevHash = old blocks hash value 
	newBlock.Hash = calculateHashValue(newBlock)  //Generate Hash value to new block

	return newBlock, nil
}

// Function to check if the new block is valid or not
func isValidBlock(newBlock,prevBlock MyBlock) bool {

	// Check if the index no of the new block is equals to prev blocks index+1
	if prevBlock.Index+1 != newBlock.Index {
		return false
	}

	// Prev blocks hash and new blocks hash cannot be equal
	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}

	// When calcualte a new hash value for new block it cannot be equal to its current hash value
	if calculateHashValue(newBlock) != newBlock.Hash {
		return false
	}

	return true
}


// Function for run the block chain in PORT-8080
func run() error {
	route := makeRouter()
	httpAddr := os.Getenv("PORT")
	log.Println("Listening on ", os.Getenv("PORT"))

	// HTTP server
	server := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        route,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Listen server in the PORT
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// Make Routers in of the Server
func makeRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", getBlockchain).Methods("GET")   //Route for the GET Method (Retriew Blocks in the Block Chain)
	router.HandleFunc("/", postBlock).Methods("POST")     //Route for POST Method (Add Blocks to the Block Chain)
	return router
}

// Method to GET BlockChain - Retriew all the Blocks as JSON Objects
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")  //Retriew all blocks as json objects
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,string(bytes))
}

// Method to POST a Block - Add a Block to blockchain as a MyBlock Object
func postBlock(w http.ResponseWriter, r *http.Request) {

	var m Message   //Message of the Block- block No

	decoder := json.NewDecoder(r.Body)

	// Decode the message
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	// Generate a block according to the message(Block No)
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BlockNo)

	// If there is an error with creating the block return the json error response
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}

	// Check blocks validity
	if isValidBlock(newBlock, Blockchain[len(Blockchain)-1]) {

		// Append new Block to Master Blockchain and assign it to a local block chain
		newBlockchain := append(Blockchain,newBlock)

		// Relace Master Block chain with the Local one
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	// Return the Json Output of the newBlock
	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

// Function for replace the master blockchain with the local block chain
func replaceChain(newBlocks []MyBlock) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

// Function to create the JSON Response
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// Main Method 
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Add the Initial Block to master Block Chain
	go func() {
		time := time.Now()
		initialBlock := MyBlock{0, time.String(), 0, "", ""}  //Initiliaze the initial block
		spew.Dump(initialBlock)
		Blockchain = append(Blockchain,initialBlock)
	}()

	// Run the Server
	log.Fatal(run())

}






