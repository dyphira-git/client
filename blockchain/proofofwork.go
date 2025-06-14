package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"math/big"
)

const maxNonce = math.MaxInt64

// currentDifficulty stores the current mining difficulty
var currentDifficulty = InitialDifficulty

// GetTargetBits returns the current mining difficulty
func GetTargetBits() int {
	return currentDifficulty
}

// SetDifficulty sets the current mining difficulty
func SetDifficulty(difficulty int) {
	if difficulty < MinDifficulty {
		difficulty = MinDifficulty
	}
	if difficulty > MaxDifficulty {
		difficulty = MaxDifficulty
	}
	currentDifficulty = difficulty
}

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-currentDifficulty))

	pow := &ProofOfWork{b, target}
	return pow
}

// prepareData prepares data for hashing
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(currentDifficulty)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	log.Printf("[Miner] Starting proof of work with target bits: %d", currentDifficulty)
	fmt.Printf("[Miner] Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			log.Printf("[Miner] Found valid proof of work - Hash: %x, Nonce: %d", hash, nonce)
			break
		} else {
			nonce++
		}

		if nonce%100000 == 0 {
			fmt.Printf("\r[Miner] Mining... Current nonce: %d", nonce)
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
