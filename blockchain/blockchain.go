package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
)

var dbFile string

func init() {
	// Get the absolute path of the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	// Set dbFile to an absolute path in the project root
	dbFile = filepath.Join(dir, "blockchain.db")
}

const blocksBucket = "blocks"
const genesisCoinbaseData = "Dyphira Genesis Block"

// Blockchain represents a blockchain
type Blockchain struct {
	tip     []byte
	DB      *bolt.DB
	mempool []*Transaction
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	// Create genesis block with 50 DYP reward
	cbtx := NewCoinbaseTx(address, genesisCoinbaseData, true, 0)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

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

	bc := Blockchain{tip, db, nil}

	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	if !dbExists() {
		log.Panic("No existing blockchain found. Create one first using CreateBlockchain(address)")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock(NewCoinbaseTx("Genesis", genesisCoinbaseData, true, 0))
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := &Blockchain{
		tip:     tip,
		DB:      db,
		mempool: make([]*Transaction, 0),
	}

	return bc
}

// AddBlock adds a mined block to the blockchain
func (bc *Blockchain) AddBlock(block *Block, transactions []*Transaction) error {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		// Check if block already exists
		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return fmt.Errorf("block already exists")
		}

		// Get the current height
		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		// Verify block height is correct
		expectedHeight := lastBlock.Height + 1
		if block.Height != expectedHeight {
			return fmt.Errorf("invalid block height: got %d, want %d", block.Height, expectedHeight)
		}

		// Verify block links to current tip
		if !bytes.Equal(block.PrevBlockHash, lastHash) {
			return fmt.Errorf("block does not link to current tip")
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Printf("Failed to add block to chain: %v", err)
			return err
		}

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			log.Printf("Failed to update chain tip: %v", err)
			return err
		}
		bc.tip = block.Hash

		return nil
	})

	if err == nil {
		// Clear the mined transactions from mempool only if the block was successfully added
		bc.ClearTransactionsFromMempool(transactions)
	}

	return err
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey *ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.DB}

	return bci
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Check if the output was spent
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// If the output address matches, it's spendable by the owner
				if strings.EqualFold(out.Address, address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// If this is not a coinbase transaction, collect all inputs that spent outputs
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount float32) (float32, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := float32(0)

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if strings.EqualFold(out.Address, address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// GetGenesisAddress returns the address that created the blockchain
func (bc *Blockchain) GetGenesisAddress() string {
	bci := bc.Iterator()

	// Iterate to the genesis block
	var genesisBlock *Block
	for {
		block := bci.Next()
		if len(block.PrevBlockHash) == 0 {
			genesisBlock = block
			break
		}
	}

	// Get the coinbase transaction from genesis block
	genesisTx := genesisBlock.Transactions[0]
	if !genesisTx.IsCoinbase() {
		return "" // This shouldn't happen in a properly initialized blockchain
	}

	// Get the recipient address from the first output
	return genesisTx.Vout[0].Address
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Printf("Database file not found at: %s\n", dbFile)
		return false
	}
	return true
}

// PrepareNewBlock creates a new block with the given transactions but doesn't mine it
func (bc *Blockchain) PrepareNewBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		lastBlock := DeserializeBlock(b.Get(lastHash))
		lastHeight = lastBlock.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
	return newBlock
}

// GetHeight returns the height of the blockchain
func (bc *Blockchain) GetHeight() int {
	var height int

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)
		height = lastBlock.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return height
}

// GetLastBlock returns the last block in the chain
func (bc *Blockchain) GetLastBlock() *Block {
	var lastBlock *Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = DeserializeBlock(blockData)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return lastBlock
}

// AddTransaction adds a new transaction to the transaction pool
func (bc *Blockchain) AddTransaction(tx *Transaction) {
	bc.mempool = append(bc.mempool, tx)
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	fmt.Println("Mining block with transactions:", transactions)
	var lastHash []byte
	var lastHeight int

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		lastBlock := DeserializeBlock(b.Get(lastHash))
		lastHeight = lastBlock.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
	newBlock.MineBlock() // Mine the block

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// GetPendingTransactions returns all transactions from the mempool
func (bc *Blockchain) GetPendingTransactions() []*Transaction {
	return bc.mempool
}

// ClearTransactionsFromMempool removes the given transactions from the mempool
func (bc *Blockchain) ClearTransactionsFromMempool(transactions []*Transaction) {

	// Create a map of transaction IDs to remove for O(1) lookup
	toRemove := make(map[string]bool)
	for _, tx := range transactions {
		if tx.ID == nil {
			continue // Skip transactions without IDs
		}
		txID := hex.EncodeToString(tx.ID)
		if txID == "" {
			continue // Skip empty transaction IDs
		}
		toRemove[txID] = true
	}

	// Filter out the transactions that were mined
	newMempool := make([]*Transaction, 0)
	for _, tx := range bc.mempool {
		if tx.ID == nil {
			continue // Skip transactions without IDs
		}
		txID := hex.EncodeToString(tx.ID)
		if txID == "" {
			continue // Skip empty transaction IDs
		}
		if !toRemove[txID] {
			newMempool = append(newMempool, tx)
		} else {
		}
	}

	bc.mempool = newMempool
}

// GetAllTransactions returns all transactions in the blockchain
func (bc *Blockchain) GetAllTransactions() []Transaction {
	var transactions []Transaction
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			transactions = append(transactions, *tx)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transactions
}

// TransactionHistoryItem represents a single transaction in the history
type TransactionHistoryItem struct {
	TxID        string  `json:"txId"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Amount      float32 `json:"amount"`
	Fee         float32 `json:"fee"`
	BlockHeight int     `json:"blockHeight"`
	Timestamp   int64   `json:"timestamp"`
	Type        string  `json:"type"` // "sent", "received", or "mining_reward"
}

// GetTransactionHistory returns the transaction history for a given address
func (bc *Blockchain) GetTransactionHistory(address string) ([]TransactionHistoryItem, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address")
	}

	history := make([]TransactionHistoryItem, 0)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			// Check if transaction is related to the address
			isInvolved := false
			txType := ""

			// Check outputs (receiving)
			for _, out := range tx.Vout {
				if strings.EqualFold(out.Address, address) {
					isInvolved = true
					txType = "received"
					break
				}
			}

			// Check inputs (sending)
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(address) {
						isInvolved = true
						txType = "sent"
						break
					}
				}
			} else if isInvolved {
				txType = "mining_reward"
			}

			if isInvolved {
				// Convert transaction ID to hex string for readability
				txIDHex := hex.EncodeToString(tx.ID)
				item := TransactionHistoryItem{
					TxID:        txIDHex,
					From:        tx.From,
					To:          tx.To,
					Amount:      tx.Amount,
					Fee:         tx.Fee,
					BlockHeight: block.Height,
					Timestamp:   block.Timestamp,
					Type:        txType,
				}
				history = append(history, item)
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	// Sort transactions by block height and timestamp
	sort.Slice(history, func(i, j int) bool {
		if history[i].BlockHeight == history[j].BlockHeight {
			return history[i].Timestamp < history[j].Timestamp
		}
		return history[i].BlockHeight < history[j].BlockHeight
	})

	return history, nil
}

// GetBalance returns the balance of an address
func (bc *Blockchain) GetBalance(address string) float32 {
	if !common.IsHexAddress(address) {
		return 0
	}

	balance := float32(0)
	unspentTXs := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if strings.EqualFold(out.Address, address) {
				balance += out.Value
			}
		}
	}

	return balance
}
