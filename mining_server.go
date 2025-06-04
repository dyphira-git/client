package main

import (
	"context"
	"encoding/hex"
	"log"

	blockchain "dyp_chain/blockchain"
	pb "dyp_chain/proto"
	"sync"
)

type miningServer struct {
	pb.UnimplementedMiningServiceServer
	blockchain *blockchain.Blockchain
	mu         sync.Mutex
}

func newMiningServer(bc *blockchain.Blockchain) *miningServer {
	return &miningServer{
		blockchain: bc,
	}
}

// MineBlock handles the mining request from clients
func (s *miningServer) MineBlock(ctx context.Context, req *pb.MineBlockRequest) (*pb.MineBlockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Starting mining process with %d transactions", len(req.Transactions))

	// Get transactions from mempool
	pendingTxs := s.blockchain.GetPendingTransactions()
	transactions := make([]*blockchain.Transaction, len(pendingTxs))
	for i, tx := range pendingTxs {
		transactions[i] = tx // Use the complete transaction from mempool
		log.Printf("Processing transaction %d:", i+1)
		log.Printf("  - ID: %x", transactions[i].ID)
		log.Printf("  - From: %s", transactions[i].From)
		log.Printf("  - To: %s", transactions[i].To)
		log.Printf("  - Amount: %d", transactions[i].Amount)
	}

	// Add mining reward transaction
	reward := blockchain.NewCoinbaseTx(req.MinerAddress, "Mining reward")
	log.Printf("Adding mining reward transaction:")
	log.Printf("  - ID: %x", reward.ID)
	log.Printf("  - To: %s", reward.To)
	log.Printf("  - Amount: %d", reward.Amount)
	transactions = append(transactions, reward)

	// Create and mine a new block
	log.Printf("Preparing new block for mining...")
	block := s.blockchain.PrepareNewBlock(transactions)
	log.Printf("Starting proof of work for block at height %d", block.Height)
	pow := blockchain.NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash
	log.Printf("Found solution - Block hash: %x, Nonce: %d", hash, nonce)

	// Add the block to the blockchain
	log.Printf("Attempting to add block to blockchain...")
	err := s.blockchain.AddBlock(block, transactions)
	if err != nil {
		log.Printf("Failed to add block to blockchain: %v", err)
		return &pb.MineBlockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	log.Printf("Successfully mined and added block %x", hash)
	return &pb.MineBlockResponse{
		Success:   true,
		BlockHash: hex.EncodeToString(hash),
	}, nil
}

// GetBlockchainStatus returns the current status of the blockchain
func (s *miningServer) GetBlockchainStatus(ctx context.Context, req *pb.BlockchainStatusRequest) (*pb.BlockchainStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tip := s.blockchain.GetLastBlock()

	return &pb.BlockchainStatusResponse{
		Height:          int32(s.blockchain.GetHeight()),
		LatestBlockHash: hex.EncodeToString(tip.Hash),
		Difficulty:      int32(16),
	}, nil
}

// GetPendingTransactions returns all pending transactions from the mempool
func (s *miningServer) GetPendingTransactions(ctx context.Context, req *pb.PendingTransactionsRequest) (*pb.PendingTransactionsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get pending transactions from the mempool
	pendingTxs := s.blockchain.GetPendingTransactions()

	// Convert blockchain transactions to proto transactions
	protoTxs := make([]*pb.Transaction, 0) // Only include valid transactions
	for _, tx := range pendingTxs {
		if tx.ID == nil {
			continue // Skip invalid transactions
		}
		protoTx := &pb.Transaction{
			From:          tx.From,
			To:            tx.To,
			Amount:        tx.Amount,
			Signature:     tx.Signature,
			TransactionId: tx.ID, // Use the new transaction_id field
		}
		protoTxs = append(protoTxs, protoTx)
	}

	return &pb.PendingTransactionsResponse{
		Transactions: protoTxs,
	}, nil
}
