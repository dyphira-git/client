package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	blockchain "dyp_chain/blockchain"
	pb "dyp_chain/proto"

	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
)

const blocksBucket = "blocks"

type miningServer struct {
	pb.UnimplementedMiningServiceServer
	blockchain  *blockchain.Blockchain
	mu          sync.Mutex
	lastAdjTime time.Time
}

// NewMiningServer creates a new mining server instance
func NewMiningServer(bc *blockchain.Blockchain) *miningServer {
	return &miningServer{
		blockchain:  bc,
		lastAdjTime: time.Now(),
	}
}

// calculateBlockSize calculates the approximate size of a block in bytes
func calculateBlockSize(block *blockchain.Block) int {
	size := 0
	// Add fixed block header size (timestamp, prevBlockHash, hash, nonce, height)
	size += 8 + len(block.PrevBlockHash) + len(block.Hash) + 4 + 4

	// Add transaction sizes
	for _, tx := range block.Transactions {
		// Add fixed transaction fields
		size += len(tx.ID) + len(tx.From) + len(tx.To) + 8 + 8 + len(tx.Signature)

		// Add inputs
		for _, in := range tx.Vin {
			size += len(in.Txid) + 4 + len(in.Signature) + len(in.PubKey)
		}

		// Add outputs
		for _, out := range tx.Vout {
			size += 4 + len(out.Address)
		}
	}

	return size
}

// GetBlockTemplate prepares a new block template for mining
func (s *miningServer) GetBlockTemplate(ctx context.Context, req *pb.BlockTemplateRequest) (*pb.BlockTemplateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !common.IsHexAddress(req.MinerAddress) {
		return nil, fmt.Errorf("invalid miner Ethereum address format")
	}

	log.Printf("[Server] Getting block template for miner: %s", req.MinerAddress)

	// Get transactions from mempool
	pendingTxs := s.blockchain.GetPendingTransactions()
	log.Printf("[Server] Found %d total transactions in mempool", len(pendingTxs))

	// Create a slice of transactions with their sizes and fees for sorting
	type txWithMetadata struct {
		tx   *blockchain.Transaction
		size int
		fee  float32
	}

	var txsMetadata []txWithMetadata

	// Calculate size and fee for each transaction
	for _, tx := range pendingTxs {
		if !tx.IsCoinbase() {
			// Calculate transaction size
			size := len(tx.ID) + len(tx.From) + len(tx.To) + 8 + 8 + len(tx.Signature)
			for _, in := range tx.Vin {
				size += len(in.Txid) + 4 + len(in.Signature) + len(in.PubKey)
			}
			for _, out := range tx.Vout {
				size += 4 + len(out.Address)
			}

			txsMetadata = append(txsMetadata, txWithMetadata{
				tx:   tx,
				size: size,
				fee:  tx.Fee,
			})
		}
	}

	// Sort transactions by fee per byte (descending) to maximize reward/size ratio
	sort.Slice(txsMetadata, func(i, j int) bool {
		feePerByteI := float64(txsMetadata[i].fee) / float64(txsMetadata[i].size)
		feePerByteJ := float64(txsMetadata[j].fee) / float64(txsMetadata[j].size)
		return feePerByteI > feePerByteJ
	})

	// Select transactions while respecting block size limit
	var selectedTxs []*blockchain.Transaction
	totalSize := 0
	totalFees := float32(0)
	realTxCount := 0

	// Reserve space for block header and coinbase transaction
	headerSize := 8 + 32 + 32 + 4 + 4 // timestamp + prevBlockHash + hash + nonce + height
	coinbaseSize := 100               // Approximate size for coinbase transaction
	remainingSize := blockchain.MaxBlockSize - headerSize - coinbaseSize

	// Select transactions that fit in the block
	for _, txMeta := range txsMetadata {
		if totalSize+txMeta.size <= remainingSize {
			selectedTxs = append(selectedTxs, txMeta.tx)
			totalSize += txMeta.size
			totalFees += txMeta.tx.Fee
			realTxCount++
			log.Printf("[Server] Selected transaction: From=%s, To=%s, Amount=%f, Fee=%f, Size=%d",
				txMeta.tx.From, txMeta.tx.To, txMeta.tx.Amount, txMeta.tx.Fee, txMeta.size)
		}
	}

	log.Printf("[Server] Selected %d transactions with total fees: %f", realTxCount, totalFees)

	// Check if this is genesis block
	var isGenesis bool
	err := s.blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			isGenesis = true
			return nil
		}
		c := b.Cursor()
		k, _ := c.First()
		isGenesis = k == nil
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check blockchain state: %v", err)
	}

	// Add mining reward transaction
	reward := blockchain.NewCoinbaseTx(req.MinerAddress, "Mining reward", isGenesis, totalFees)
	selectedTxs = append(selectedTxs, reward)

	// Create a block template
	block := s.blockchain.PrepareNewBlock(selectedTxs)

	// Double check final block size
	blockSize := calculateBlockSize(block)
	if blockSize > blockchain.MaxBlockSize {
		log.Printf("[Server] Block size %d exceeds maximum %d bytes", blockSize, blockchain.MaxBlockSize)
		return nil, fmt.Errorf("block size exceeds maximum allowed size")
	}

	// Adjust difficulty if needed
	lastBlock := s.blockchain.GetLastBlock()
	if lastBlock != nil && !isGenesis {
		if block.Height%blockchain.DifficultyAdjustmentInterval == 0 {
			actualTimespan := time.Since(s.lastAdjTime)
			newDifficulty := blockchain.CalculateNextDifficulty(blockchain.GetTargetBits(), actualTimespan)
			blockchain.SetDifficulty(newDifficulty)
			s.lastAdjTime = time.Now()
			log.Printf("[Server] Adjusted difficulty to %d", newDifficulty)
		}
	}

	log.Printf("[Server] Prepared block template: Height=%d, PrevHash=%x, Size=%d bytes",
		block.Height, block.PrevBlockHash, blockSize)

	// Convert block to protobuf format
	pbBlock := &pb.Block{
		Timestamp:     block.Timestamp,
		PrevBlockHash: block.PrevBlockHash,
		Height:        int32(block.Height),
		Transactions:  make([]*pb.Transaction, len(block.Transactions)),
	}

	// Convert transactions
	for i, tx := range block.Transactions {
		pbTx := &pb.Transaction{
			From:          tx.From,
			To:            tx.To,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
			TransactionId: tx.ID,
			Signature:     tx.Signature,
			Vin:           make([]*pb.TXInput, len(tx.Vin)),
			Vout:          make([]*pb.TXOutput, len(tx.Vout)),
		}

		if tx.IsCoinbase() {
			if isGenesis {
				log.Printf("[Server] Converting genesis coinbase transaction for miner: %s, Amount=%f DYP (50 DYP + %f fees)",
					tx.To, tx.Amount, totalFees)
			} else {
				log.Printf("[Server] Converting coinbase transaction for miner: %s, Amount=%f DYP (50 DYP + %f fees)",
					tx.To, tx.Amount, totalFees)
			}
			pbTx.Vin = []*pb.TXInput{
				{
					Txid:      []byte{},
					Vout:      -1,
					Signature: tx.Signature,
					PubKey:    []byte("Mining reward"),
				},
			}
			pbTx.Vout = []*pb.TXOutput{
				{
					Value:   tx.Amount,
					Address: tx.To,
				},
			}
		} else {
			for j, vin := range tx.Vin {
				pbTx.Vin[j] = &pb.TXInput{
					Txid:      vin.Txid,
					Vout:      int32(vin.Vout),
					Signature: vin.Signature,
					PubKey:    vin.PubKey,
				}
			}
			for j, vout := range tx.Vout {
				pbTx.Vout[j] = &pb.TXOutput{
					Value:   vout.Value,
					Address: vout.Address,
				}
			}
		}
		pbBlock.Transactions[i] = pbTx
	}

	return &pb.BlockTemplateResponse{
		Block:      pbBlock,
		Difficulty: int32(blockchain.GetTargetBits()),
	}, nil
}

// SubmitBlock handles the submission of a mined block
func (s *miningServer) SubmitBlock(ctx context.Context, req *pb.SubmitBlockRequest) (*pb.SubmitBlockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[Server] Received block submission: Height=%d, Hash=%x", req.Block.Height, req.BlockHash)

	// Convert protobuf block to blockchain.Block
	transactions := make([]*blockchain.Transaction, len(req.Block.Transactions))
	realTxCount := 0
	totalFees := float32(0)

	for i, tx := range req.Block.Transactions {
		transactions[i] = &blockchain.Transaction{
			ID:        tx.TransactionId,
			From:      tx.From,
			To:        tx.To,
			Amount:    tx.Amount,
			Fee:       tx.Fee,
			Signature: tx.Signature,
			Vin:       make([]blockchain.TXInput, len(tx.Vin)),
			Vout:      make([]blockchain.TXOutput, len(tx.Vout)),
		}

		// Convert inputs
		for j, vin := range tx.Vin {
			transactions[i].Vin[j] = blockchain.TXInput{
				Txid:      vin.Txid,
				Vout:      int(vin.Vout),
				Signature: vin.Signature,
				PubKey:    vin.PubKey,
			}
		}

		// Convert outputs
		for j, vout := range tx.Vout {
			transactions[i].Vout[j] = blockchain.TXOutput{
				Value:   vout.Value,
				Address: vout.Address,
			}
		}

		if tx.From == "coinbase" {
			log.Printf("[Server] Processing coinbase transaction for miner: %s, Amount=%f", tx.To, tx.Amount)
		} else {
			realTxCount++
			totalFees += tx.Fee
			log.Printf("[Server] Processing regular transaction: From=%s, To=%s, Amount=%f, Fee=%f", tx.From, tx.To, tx.Amount, tx.Fee)
		}
	}
	log.Printf("[Server] Block contains %d real transactions and 1 coinbase transaction, total fees: %f", realTxCount, totalFees)

	// Create the block without mining it
	block := &blockchain.Block{
		Timestamp:     req.Block.Timestamp,
		Transactions:  transactions,
		PrevBlockHash: req.Block.PrevBlockHash,
		Hash:          req.BlockHash,
		Nonce:         int(req.Nonce),
		Height:        int(req.Block.Height),
	}

	// Verify the proof of work
	pow := blockchain.NewProofOfWork(block)
	if !pow.Validate() {
		log.Printf("[Server] Block validation failed: invalid proof of work. Hash=%x, Nonce=%d", req.BlockHash, req.Nonce)
		return &pb.SubmitBlockResponse{
			Success:      false,
			ErrorMessage: "invalid proof of work",
		}, nil
	}
	log.Printf("[Server] Block proof of work validation successful")

	// Add the block to the blockchain
	err := s.blockchain.AddBlock(block, transactions)
	if err != nil {
		log.Printf("[Server] Failed to add block to chain: %v", err)
		return &pb.SubmitBlockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	log.Printf("[Server] Successfully added block to chain: Height=%d, Hash=%x", block.Height, block.Hash)
	return &pb.SubmitBlockResponse{
		Success:   true,
		BlockHash: hex.EncodeToString(req.BlockHash),
	}, nil
}

// GetBlockchainStatus returns the current status of the blockchain
func (s *miningServer) GetBlockchainStatus(ctx context.Context, req *pb.BlockchainStatusRequest) (*pb.BlockchainStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tip := s.blockchain.GetLastBlock()
	log.Printf("[Server] Current blockchain status: Height=%d, TipHash=%x", tip.Height, tip.Hash)

	return &pb.BlockchainStatusResponse{
		Height:          int32(s.blockchain.GetHeight()),
		LatestBlockHash: hex.EncodeToString(tip.Hash),
		Difficulty:      int32(blockchain.GetTargetBits()),
	}, nil
}

// GetPendingTransactions returns all pending transactions from the mempool
func (s *miningServer) GetPendingTransactions(ctx context.Context, req *pb.PendingTransactionsRequest) (*pb.PendingTransactionsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pendingTxs := s.blockchain.GetPendingTransactions()
	pbTxs := make([]*pb.Transaction, len(pendingTxs))

	for i, tx := range pendingTxs {
		pbTx := &pb.Transaction{
			From:          tx.From,
			To:            tx.To,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
			TransactionId: tx.ID,
			Signature:     tx.Signature,
		}
		pbTxs[i] = pbTx
	}

	return &pb.PendingTransactionsResponse{
		Transactions: pbTxs,
	}, nil
}
