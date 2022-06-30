package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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

// 看看是否满足coinbase的特征
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
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
// 我的钱
func (in *TXInput) CanUnlockOutWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// TODO: Output  能否被解锁，这俩函数一个不就够了嘛
// 别人转给我的钱
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubLey == unlockingData
}
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	// A coinbase Transaction has only one input. Its input is empty and its Vout equals to -1.
	// The input doesnt store a script in ScriptSig. Arbitrary data is stored there.
	txin := TXInput{[]byte{}, -1, data}
	// subsidy is the amount of reward.
	txout := TXOutput{subsidy, to}

	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

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
// a general transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: NOT enogh funds")
	}

	for txid, outs := range validOutputs {

		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Fatal("Decode err")
		}
		// 把所有的outputs连起来作为输入
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}
	// 第一个out是给谁
	outputs = append(outputs, TXOutput{amount, to})
	// 还有多余的就找零
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	// 打包成一个交易
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// 找到所有的未花费输出，并且确定他们有足够的价值
