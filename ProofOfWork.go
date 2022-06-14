package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// the difficulty
// the big, the hard
const targetBits = 12

var (
	maxNonce = math.MaxInt64
)

// target difficulty
type ProofOfWork struct {
	block *Block
	// check if it is less than
	target *big.Int
}

// get target proof value
func NewProofOfWork(b *Block) *ProofOfWork {

	target := big.NewInt(1)

	//  Lsh sets target = target << n and returns target.
	target.Lsh(target, uint((256 - targetBits)))

	// get a pow based on given target difficulty
	pow := &ProofOfWork{b, target}

	return pow
}

// block --> ProofOfWork --> prepareDataOfTheBlock
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.TimeStamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	// while(nouce < maxNonce)
	for nonce < maxNonce {
		// get different hash for different nonce
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		// covverted into a big integer
		hashInt.SetBytes(hash[:])

		// the hash of the current nonce compared to the pow.target
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			// a new nonce
			nonce++
		}
	}
	fmt.Printf("\n nonce=[%d]\n\n", nonce)
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	// get hash
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	// 就比较个大小
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
