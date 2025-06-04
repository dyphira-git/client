package main

import (
	"context"
	"encoding/hex"

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

	// Convert proto transactions to blockchain transactions
	transactions := make([]*blockchain.Transaction, len(req.Transactions))
	for i, tx := range req.Transactions {
		transactions[i] = &blockchain.Transaction{
			From:      tx.From,
			To:        tx.To,
			Amount:    tx.Amount,
			Signature: tx.Signature,
		}
	}

	// Create and mine a new block
	block := s.blockchain.PrepareNewBlock(transactions)
	pow := blockchain.NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	// Add the block to the blockchain
	err := s.blockchain.AddBlock(block, transactions)
	if err != nil {
		return &pb.MineBlockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Reward the miner
	reward := blockchain.NewCoinbaseTx(req.MinerAddress, "Mining reward")
	s.blockchain.AddTransaction(reward)

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
