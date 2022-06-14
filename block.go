package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strconv"
	"time"
)

type Block struct {
	// 1. when the block is created
	TimeStamp int64
	// 2. information contained in the block
	Data []byte
	// 3. the hash of pervious block ---用来遍历
	PrevBlockHash []byte
	// 4. the hash of the block
	Hash []byte
	// the times the block mine the approiate hash
	Nonce int
}
//
func (b *Block) SetHash() {
	// 1. int64 ->  str  -> []byte; in the given base of 10
	timestamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
	// 2.concatenate elements of slices of slices tp get a slice
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	// 3. returns the SHA256 checksum of the data.
	hash := sha256.Sum256(headers)
	// 4. reflect.TypeOf(hash[:]) ----> []byte
	b.Hash = hash[:]
}

// creation of a block
func NewBlock(data string, prevBlockHash []byte) *Block {
	// init of a block
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	// when creating a new block, produce a pow for the block
	pow := NewProofOfWork(block)
	// get the hash less than pow.target
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// b.Serialize() ? 将block序列化为直接数组
func (b *Block) Serialize() []byte {
	// 1. a buffer store serialized data
	var result bytes.Buffer
	// 2. get a encoder and then eocode block
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	// 3. return the serialized result of block
	return result.Bytes()
}

// byte -> Block
// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
