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
	"syscall"
	"time"

	pb "dyp_chain/proto"

	"bytes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MiningClient struct {
	client pb.MiningServiceClient
	conn   *grpc.ClientConn
}

func NewMiningClient(address string) (*MiningClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewMiningServiceClient(conn)
	return &MiningClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *MiningClient) Close() {
	if c.conn != nil {
		c.conn.Close()
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
		data := prepareData(block, nonce)
		hash := sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		hashesComputed++
		if hashesComputed%1000000 == 0 {
			elapsed := time.Since(startTime)
			hashrate := float64(hashesComputed) / elapsed.Seconds()
			log.Printf("[Miner] Mining progress: Nonce=%d, Hashrate=%.2f H/s", nonce, hashrate)
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
	log.Printf("[Miner] Failed to find solution within nonce range")
	return nonce, nil
}

// prepareData prepares block data for hashing
func prepareData(block *pb.Block, nonce int32) []byte {
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
	binary.BigEndian.PutUint64(targetBits, uint64(16)) // Using constant targetBits=16

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

func (c *MiningClient) StartMining(minerAddress string) {
	log.Printf("[Miner] Starting mining operations for address: %s", minerAddress)

	// Create a channel to handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start mining in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[Miner] Mining operations stopped")
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

				// Get pending transactions count
				pendingTxs, err := c.client.GetPendingTransactions(ctx, &pb.PendingTransactionsRequest{})
				if err != nil {
					log.Printf("[Miner] Failed to get pending transactions: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Count non-coinbase transactions
				realTxCount := 0
				for _, tx := range pendingTxs.Transactions {
					if tx.From != "coinbase" {
						realTxCount++
						log.Printf("[Miner] Found pending transaction: From=%s, To=%s, Amount=%f",
							tx.From, tx.To, tx.Amount)
					}
				}

				if realTxCount == 0 {
					log.Printf("[Miner] No real transactions to mine, waiting for new transactions...")
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("[Miner] Found %d real transactions ready for mining", realTxCount)

				// Get block template
				template, err := c.client.GetBlockTemplate(ctx, &pb.BlockTemplateRequest{
					MinerAddress: minerAddress,
				})
				if err != nil {
					if strings.Contains(err.Error(), "no transactions available") {
						log.Printf("[Miner] Server reports no transactions available")
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
					log.Printf("[Miner] Mining attempt failed, retrying with new template")
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

				resp, err := c.client.SubmitBlock(ctx, submitReq)
				if err != nil {
					log.Printf("[Miner] Failed to submit block: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				if resp.Success {
					log.Printf("[Miner] Block accepted by network: Hash=%s", resp.BlockHash)
					// Add a delay after successful mining to allow the network to sync
					time.Sleep(2 * time.Second)
				} else {
					log.Printf("[Miner] Block rejected: %s", resp.ErrorMessage)
					if strings.Contains(resp.ErrorMessage, "invalid proof of work") {
						log.Printf("[Miner] Invalid proof of work, retrying with new template")
						continue
					}
				}

				// Small delay between mining attempts
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Printf("[Miner] Received shutdown signal, cleaning up...")
	cancel()
	time.Sleep(1 * time.Second) // Give a moment for cleanup
}
