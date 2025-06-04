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

	"github.com/boltdb/bolt"
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
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

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

	cbtx := NewCoinbaseTx(address, genesisCoinbaseData)
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
			genesis := NewGenesisBlock(NewCoinbaseTx("Genesis", genesisCoinbaseData))
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
	log.Printf("Attempting to add block to chain - Height: %d, Hash: %x", block.Height, block.Hash)
	log.Printf("Block contains %d transactions:", len(block.Transactions))
	for i, tx := range block.Transactions {
		log.Printf("  Transaction %d:", i+1)
		log.Printf("    - ID: %x", tx.ID)
		log.Printf("    - From: %s", tx.From)
		log.Printf("    - To: %s", tx.To)
		log.Printf("    - Amount: %d", tx.Amount)
	}

	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return fmt.Errorf("block already exists")
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Printf("Failed to add block to chain: %v", err)
			return err
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Printf("Failed to update chain tip: %v", err)
				return err
			}
			bc.tip = block.Hash
			log.Printf("Successfully updated chain tip to new block")
		}

		return nil
	})

	if err == nil {
		log.Printf("Successfully added block to chain")
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
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
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
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	log.Printf("Starting FindUnspentTransactions for pubKeyHash: %x", pubKeyHash)
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	blockHeight := 0
	for {
		block := bci.Next()
		log.Printf("Processing block at height %d, hash: %x", blockHeight, block.Hash)

		for txIndex, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			log.Printf("  Analyzing transaction %d/%d - ID: %s", txIndex+1, len(block.Transactions), txID)

		Outputs:
			for outIdx, out := range tx.Vout {
				log.Printf("    Checking output %d (value: %d)", outIdx, out.Value)

				// Check if the output was spent
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							log.Printf("      Output %d is already spent", outIdx)
							continue Outputs
						}
					}
				}

				// If the output is locked with the provided pubKeyHash, it's spendable by the owner
				if out.IsLockedWithKey(pubKeyHash) {
					log.Printf("      Found unspent output %d belonging to pubKeyHash", outIdx)
					unspentTXs = append(unspentTXs, *tx)
				} else {
					log.Printf("      Output %d is locked with different pubKeyHash", outIdx)
				}
			}

			// If this is not a coinbase transaction, collect all inputs that spent outputs
			if !tx.IsCoinbase() {
				log.Printf("    Processing inputs for non-coinbase transaction")
				for inIdx, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						log.Printf("      Input %d spends output %d from transaction %s", inIdx, in.Vout, inTxID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			} else {
				log.Printf("    Skipping inputs (coinbase transaction)")
			}
		}

		if len(block.PrevBlockHash) == 0 {
			log.Printf("Reached genesis block, stopping iteration")
			break
		}
		blockHeight++
	}

	log.Printf("Found %d unspent transactions", len(unspentTXs))
	return unspentTXs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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
	pubKeyHash := genesisTx.Vout[0].PubKeyHash
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := string(Base58Encode(fullPayload))

	return address
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
	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run()

	newBlock.Hash = hash
	newBlock.Nonce = nonce

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

	log.Printf("Starting to collect all transactions from blockchain")
	blockHeight := 0
	for {
		block := bci.Next()
		log.Printf("Processing block at height %d, hash: %x", blockHeight, block.Hash)

		for _, tx := range block.Transactions {
			transactions = append(transactions, *tx)
			log.Printf("  Added transaction: ID=%x, From=%s, To=%s, Amount=%d",
				tx.ID, tx.From, tx.To, tx.Amount)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
		blockHeight++
	}

	log.Printf("Found total %d transactions in blockchain", len(transactions))
	return transactions
}

// TransactionHistoryItem represents a single transaction in the history
type TransactionHistoryItem struct {
	TxID        string `json:"txId"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      int64  `json:"amount"`
	BlockHeight int    `json:"blockHeight"`
	Timestamp   int64  `json:"timestamp"`
	Type        string `json:"type"` // "sent", "received", or "mining_reward"
}

// GetTransactionHistory returns the transaction history for a given address
func (bc *Blockchain) GetTransactionHistory(address string) ([]TransactionHistoryItem, error) {
	if !ValidateAddress(address) {
		return nil, fmt.Errorf("invalid address")
	}

	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

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
				if out.IsLockedWithKey(pubKeyHash) {
					isInvolved = true
					txType = "received"
					break
				}
			}

			// Check inputs (sending)
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
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

	return history, nil
}
