package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)
const subsidy = 10
type Transaction struct {
	ID []byte
	// Inputs of a new transaction reference outputs of a previous transaction
	Vin []TXInput
	// Outputs are where coins are actually stored
	Vout []TXOutput
}

type TXOutput struct {
	// outputs store coins
	Value int
	// puzzle locks coins
	// an arbitrary string , user defined wallet address
	ScriptPubLey string
}
type TXInput struct {
	// Txid stores the ID of such transaction
	Txid []byte
	// Vout stores an index of an output in the transaction
	Vout int
	// ScriptSig is a scrpt which provides data to be used in an output's ScripPubKey.
	// to unlock the referenced output
	// the mechanism that guarantees that users can't spend coins belonging to other people
	ScriptSig string
}
// TODO: input 去解锁，能解开才可以将output引用为input
func (in *TXInput) CanUnlockOutWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
// TODO: Output  能否被解锁，这俩函数一个不就够了嘛
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubLey == unlockingData
}
func NewCoinbaseTX(to, data string) * Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	// A coinbase Transaction has only one input. Its input is empty and its Vout equals to -1.
	// The input doesnt store a script in ScriptSig. Arbitrary data is stored there.
	txin := TXInput{[]byte{}, -1, data}
	// subsidy is the amount of reward.
	txout := TXOutput{subsidy, to}

	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.setID()

	return &tx
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	//
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)

	if err != nil {
		log.Panic(err)
	}
	// 转化为byte然后再hash..
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}


// 一种通用的普通交易
func NewUTXOTransaction(from, to string, amout int, bc *Blockchain) *Transaction {
}
// 找到所有的未花费输出，并且确定他们有足够的价值

