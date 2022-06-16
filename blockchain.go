package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)
// 数据库名和表名
const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "2022/6/16 星期四 晴天 休班"
type Blockchain struct {
	// tip---the hash of the last block
	tip []byte
	// a DB connection
	db *bolt.DB
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction{
	var unspentTXs []Transaction
	// 创建一个map,键是string，值是[]int
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		// 拿到每个block
		block := bci.Next()
		// 拿到block里的每个transaction
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		// 拿到每个transaction里的output
		for outIdx, out := range tx.Vout {
			if spentTXOs[txID] != nil {
				for _, spentOut := range spentTXOs[txID] {
					if spentOut == outIdx {
						continue Outputs
				 }
				}
			}
		}

		if out.CanBeUnlockedWith(address) {

		}


		}
	}
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
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([] *Transaction{coinbase}, []byte{})
}


// 1. 将创世区块(hash:serialize())和最后一个块的哈希(l:hash)，存储下来
// 返回新建好的区块链
func NewBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip, db}
	return &bc
}


// create a blockchain in db
// the address will receive the reward for mining the genesis block
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		// 1. new coinbase transaction
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		// 2. new genesis Block
		genesis := NewGenesisBlock(cbtx)
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		// 3.store geneis block
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		// 4. store the hash of the last block
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip, db}
	return &bc
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
