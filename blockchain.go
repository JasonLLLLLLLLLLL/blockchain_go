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

// 找到所有能被当前地址解锁的transaction
// The function returns a list of transactions containing unspent outputs.
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
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

			// TODO:没看懂这个循环
		Outputs:
			// 拿到每个transaction里的output
			for outIdx, out := range tx.Vout {
				// ??spentTXOs不是空的嘛
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// 看看有没有个输出能否被当前地址解锁，可以的话就把整个transaction添加到数组当中去
				// 说明是别人转给这个address的钱
				if out.CanBeUnlockedWith(address) {
					// ??
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			// 看看这个transaction是不是coinbase

			if tx.IsCoinbase() == false {
				// 遍历他的输入
				for _, in := range tx.Vin {
					// 当前这个地址能否引用
					if in.CanUnlockOutWith(address) {
						//
						inTxID := hex.EncodeToString(in.Txid)
						//
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		if len(block.PrevBlockHash)  == 0 {
			break
			}
		}
	}
	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	// 遍历每个transaction
	for _, tx := range unspentTransactions {
		// 遍历每个transation里的每个out
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
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
	newBlock := NewBlock(transactions, lastHash)
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


// Find all unspent outputs and ensure that they store enough value

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _,tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			// 不应该一定是
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				// 算钱
				accumulated += out.Value
				// 把Id都存储起来
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
// 刚刚够钱就结束了
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
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
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 1. 将创世区块(hash:serialize())和最后一个块的哈希(l:hash)，存储下来
// 返回新建好的区块链
func NewBlockchain(address string) *Blockchain {
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
