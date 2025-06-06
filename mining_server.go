package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
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
	blockchain *blockchain.Blockchain
	mu         sync.Mutex
}

func newMiningServer(bc *blockchain.Blockchain) *miningServer {
	return &miningServer{
		blockchain: bc,
	}
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

	// Check if there are any non-coinbase transactions
	hasRealTransactions := false
	realTxCount := 0
	totalFees := float32(0)
	for _, tx := range pendingTxs {
		if !tx.IsCoinbase() {
			hasRealTransactions = true
			realTxCount++
			totalFees += tx.Fee
			log.Printf("[Server] Found real transaction: From=%s, To=%s, Amount=%f, Fee=%f", tx.From, tx.To, tx.Amount, tx.Fee)
		}
	}

	if !hasRealTransactions {
		log.Printf("[Server] No real transactions found in mempool")
		return nil, fmt.Errorf("no transactions available for mining")
	}
	log.Printf("[Server] Found %d real transactions to mine, total fees: %f", realTxCount, totalFees)

	transactions := make([]*blockchain.Transaction, len(pendingTxs))
	copy(transactions, pendingTxs)

	// Check if this is genesis block by checking if any blocks exist
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
	transactions = append(transactions, reward)
	if isGenesis {
		log.Printf("[Server] Added genesis coinbase reward transaction for miner: %s, reward: %f DYP", req.MinerAddress, reward.Amount)
	} else {
		log.Printf("[Server] Added coinbase transaction with fees for miner: %s, fees: %f", req.MinerAddress, reward.Amount)
	}

	// Create a block template
	block := s.blockchain.PrepareNewBlock(transactions)
	log.Printf("[Server] Prepared block template: Height=%d, PrevHash=%x", block.Height, block.PrevBlockHash)

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

		// Convert Vin and Vout to protobuf format
		if tx.IsCoinbase() {
			// For coinbase transactions, we already have the basic fields set
			if isGenesis {
				log.Printf("[Server] Converting genesis coinbase transaction for miner: %s, Amount=%f DYP", tx.To, tx.Amount)
			} else {
				log.Printf("[Server] Converting coinbase transaction with fees for miner: %s, Amount=%f", tx.To, tx.Amount)
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
			// For regular transactions, include Vin and Vout
			log.Printf("[Server] Converting regular transaction: From=%s, To=%s, Amount=%f, Fee=%f", tx.From, tx.To, tx.Amount, tx.Fee)

			// Convert inputs
			for j, vin := range tx.Vin {
				pbTx.Vin[j] = &pb.TXInput{
					Txid:      vin.Txid,
					Vout:      int32(vin.Vout),
					Signature: vin.Signature,
					PubKey:    vin.PubKey,
				}
				log.Printf("[Server] Input %d: TxID=%x, Vout=%d", j, vin.Txid, vin.Vout)
			}

			// Convert outputs
			for j, vout := range tx.Vout {
				pbTx.Vout[j] = &pb.TXOutput{
					Value:   vout.Value,
					Address: vout.Address,
				}
				log.Printf("[Server] Output %d: Value=%f", j, vout.Value)
			}
		}

		pbBlock.Transactions[i] = pbTx
	}

	log.Printf("[Server] Sending block template with %d total transactions", len(pbBlock.Transactions))
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
		// Create a new transaction with proper Vin and Vout
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
			log.Printf("[Server] Transaction has %d inputs and %d outputs", len(tx.Vin), len(tx.Vout))
		}
	}
	log.Printf("[Server] Block contains %d real transactions and 1 coinbase transaction, total fees: %f", realTxCount, totalFees)

	// Get the current height from the blockchain
	currentHeight := s.blockchain.GetHeight()
	newHeight := currentHeight + 1

	block := &blockchain.Block{
		Timestamp:     time.Now().Unix(), // Set current timestamp
		Transactions:  transactions,
		PrevBlockHash: req.Block.PrevBlockHash,
		Hash:          req.BlockHash,
		Nonce:         int(req.Nonce),
		Height:        newHeight, // Set proper height
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
	log.Printf("[Server] Getting pending transactions, found %d total", len(pendingTxs))

	pbTxs := make([]*pb.Transaction, len(pendingTxs))
	realTxCount := 0
	for i, tx := range pendingTxs {
		pbTxs[i] = &pb.Transaction{
			From:          tx.From,
			To:            tx.To,
			Amount:        tx.Amount,
			TransactionId: tx.ID,
			Signature:     tx.Signature,
		}
		if tx.From != "coinbase" {
			realTxCount++
			log.Printf("[Server] Pending transaction %d: From=%s, To=%s, Amount=%f", i, tx.From, tx.To, tx.Amount)
		}
	}
	log.Printf("[Server] Returning %d real transactions from mempool", realTxCount)

	return &pb.PendingTransactionsResponse{
		Transactions: pbTxs,
	}, nil
}
