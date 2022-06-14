package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)
// 数据库名和表名
const dbFile = "blockchain.db"
const blocksBucket = "blocks"

type Blockchain struct {
	// tip---the hash of the last block
	tip []byte
	// a DB connection
	db *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}


//
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte
	// 1.0 ** view : read-only transactions
	err := bc.db.View(func(tx *bolt.Tx) error {
		// 1.1 get the bucket
		b := tx.Bucket([]byte(blocksBucket))
		// 1.2 get the last block hash from the bucket
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// 2. mine a new block after the last block
	newBlock := NewBlock(data, lastHash)
	// 3. a write transaction
	err = bc.db.Update(func(tx *bolt.Tx) error {
		// 3.1 get the bucket
		b := tx.Bucket([]byte(blocksBucket))
		// 3.2 store (hash : block)
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		// 3.3 update the hash of last block
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		// 3.3 update the hash of the blockchain

		bc.tip = newBlock.Hash

		return nil
	})
}

// Iterator ...
func (bc *Blockchain) Iterator() *BlockchainIterator {
	// start iterating from the last of the hash
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block
	// 套个view
	err := i.db.View(func(tx *bolt.Tx) error {
		// 表?
		b := tx.Bucket([]byte(blocksBucket))
		// 调用区块的hash
		encodedBlock := b.Get(i.currentHash)
		// 反序列化
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	// i更新为前一个区块
	i.currentHash = block.PrevBlockHash
	return block
}

// creation of the first block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}


// 1. 将创世区块(hash:serialize())和最后一个块的哈希(l:hash)，存储下来
// init of a blockchain with genesis block
func NewBlockchain() *Blockchain {
	var tip []byte
	// open a boltdb file
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		// 1. bucket storing our blocks
		b := tx.Bucket([]byte(blocksBucket))
		// 2.0 if bucket not exist
		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			// 2.1 generate the genesis block
			genesis := NewGenesisBlock()
			// 2.2 create the bucket
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			// 2.3 store the genesis block
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 2.4 [l:last block hash of the chain - genesis.Hash]
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			//

			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}
