package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "dyp_chain/proto"

	"bytes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MiningClient handles the mining operations
type MiningClient struct {
	client       pb.MiningServiceClient
	minerAddress string
	conn         *grpc.ClientConn
	stopMining   chan struct{}
	mu           sync.Mutex
	isMining     bool
}

// NewMiningClient creates a new mining client
func NewMiningClient(address string) (*MiningClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &MiningClient{
		client:       pb.NewMiningServiceClient(conn),
		minerAddress: address,
		conn:         conn,
		stopMining:   make(chan struct{}),
	}, nil
}

// StopMining signals the mining operation to stop
func (c *MiningClient) StopMining() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isMining {
		close(c.stopMining)
		c.isMining = false
		c.stopMining = make(chan struct{}) // Create new channel for next mining session
	}
}

// performProofOfWork performs the mining computation locally
func (c *MiningClient) performProofOfWork(block *pb.Block, difficulty int32) (int32, []byte) {
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	nonce := int32(0)
	maxNonce := math.MaxInt32

	log.Printf("[Miner] Starting proof of work: Height=%d, Difficulty=%d, PrevHash=%x",
		block.Height, difficulty, block.PrevBlockHash)

	realTxCount := 0
	for _, tx := range block.Transactions {
		if tx.From != "coinbase" {
			realTxCount++
		}
	}
	log.Printf("[Miner] Mining block with %d real transactions", realTxCount)

	startTime := time.Now()
	hashesComputed := 0

	for nonce < int32(maxNonce) {
		select {
		case <-c.stopMining:
			log.Printf("[Miner] Mining operation cancelled")
			return 0, nil
		default:
			data := prepareData(block, nonce, difficulty)
			hash := sha256.Sum256(data)
			hashInt.SetBytes(hash[:])

			hashesComputed++
			if hashesComputed%1000000 == 0 {
				// Check blockchain status to see if we should continue mining
				status, err := c.client.GetBlockchainStatus(context.Background(), &pb.BlockchainStatusRequest{})
				if err == nil && int(status.Height) >= int(block.Height) {
					log.Printf("[Miner] Stopping mining - new block already found at height %d", status.Height)
					return 0, nil
				}

			}

			if hashInt.Cmp(target) == -1 {
				elapsed := time.Since(startTime)
				hashrate := float64(hashesComputed) / elapsed.Seconds()
				log.Printf("[Miner] Found solution! Nonce=%d, Hash=%x, Time=%.2fs, Hashrate=%.2f H/s",
					nonce, hash, elapsed.Seconds(), hashrate)
				return nonce, hash[:]
			}

			nonce++
		}
	}
	log.Printf("[Miner] Failed to find solution within nonce range")
	return nonce, nil
}

// prepareData prepares block data for hashing
func prepareData(block *pb.Block, nonce int32, difficulty int32) []byte {
	// Hash transactions
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.TransactionId)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	// Convert numbers to hex format
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(block.Timestamp))

	targetBits := make([]byte, 8)
	binary.BigEndian.PutUint64(targetBits, uint64(difficulty)) // Use dynamic difficulty from template

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, uint64(nonce))

	// Join all data in the correct order
	data := bytes.Join(
		[][]byte{
			block.PrevBlockHash,
			txHash[:],
			timestamp,
			targetBits,
			nonceBytes,
		},
		[]byte{},
	)

	return data
}

// StartMining starts the mining operation
func (c *MiningClient) StartMining(minerAddress string) {
	log.Printf("[Miner] Starting mining operations for address: %s", minerAddress)

	// Create a channel to handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.mu.Lock()
	c.isMining = true
	c.mu.Unlock()

	// Start mining in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[Miner] Mining operations stopped")
				return
			case <-c.stopMining:
				log.Println("[Miner] Mining operation cancelled")
				return
			default:
				// Get blockchain status
				status, err := c.client.GetBlockchainStatus(ctx, &pb.BlockchainStatusRequest{})
				if err != nil {
					log.Printf("[Miner] Failed to get blockchain status: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("[Miner] Current blockchain: Height=%d, Latest=%s, Difficulty=%d",
					status.Height, status.LatestBlockHash, status.Difficulty)

				// Get block template
				template, err := c.client.GetBlockTemplate(ctx, &pb.BlockTemplateRequest{
					MinerAddress: minerAddress,
				})
				if err != nil {
					if strings.Contains(err.Error(), "block size exceeds maximum") {
						log.Printf("[Miner] Block size exceeds maximum allowed size")
						time.Sleep(5 * time.Second)
						continue
					}
					log.Printf("[Miner] Failed to get block template: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Count real transactions in template (excluding coinbase)
				realTxInTemplate := 0
				for _, tx := range template.Block.Transactions {
					if tx.From != "coinbase" {
						realTxInTemplate++
					}
				}

				log.Printf("[Miner] Received block template: Height=%d, PrevHash=%x, Transactions=%d",
					template.Block.Height, template.Block.PrevBlockHash, realTxInTemplate)

				// Perform proof of work locally
				nonce, blockHash := c.performProofOfWork(template.Block, template.Difficulty)
				if blockHash == nil {
					log.Printf("[Miner] Mining attempt cancelled or failed, retrying with new template")
					continue
				}

				// Check if the block height is still valid
				currentStatus, err := c.client.GetBlockchainStatus(ctx, &pb.BlockchainStatusRequest{})
				if err == nil && currentStatus.Height >= template.Block.Height {
					log.Printf("[Miner] Block at height %d already exists, skipping submission", template.Block.Height)
					continue
				}

				log.Printf("[Miner] Submitting block: Height=%d, Hash=%x, Nonce=%d",
					template.Block.Height, blockHash, nonce)

				// Submit the mined block
				submitReq := &pb.SubmitBlockRequest{
					Block:     template.Block,
					BlockHash: blockHash,
					Nonce:     nonce,
				}

				submitResp, err := c.client.SubmitBlock(ctx, submitReq)
				if err != nil {
					log.Printf("[Miner] Failed to submit block: %v", err)
					continue
				}

				if !submitResp.Success {
					log.Printf("[Miner] Block submission failed: %s", submitResp.ErrorMessage)
					continue
				}

				log.Printf("[Miner] Successfully submitted block")
			}
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("[Miner] Received interrupt signal, shutting down...")

	// Stop mining
	c.StopMining()

	cancel()
	time.Sleep(1 * time.Second) // Give a moment for cleanup
}

// Close closes the client connection
func (c *MiningClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
