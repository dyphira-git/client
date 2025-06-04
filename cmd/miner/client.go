package main

import (
	"context"
	pb "dyp_chain/proto"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func (c *MiningClient) StartMining(minerAddress string) {
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
				log.Println("Mining stopped")
				return
			default:
				// Get blockchain status
				status, err := c.client.GetBlockchainStatus(ctx, &pb.BlockchainStatusRequest{})
				if err != nil {
					log.Printf("Failed to get blockchain status: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("Blockchain status: Height=%d, Latest Block=%s, Difficulty=%d",
					status.Height, status.LatestBlockHash, status.Difficulty)

				// Get pending transactions
				pendingTxs, err := c.client.GetPendingTransactions(ctx, &pb.PendingTransactionsRequest{})
				if err != nil {
					log.Printf("Failed to get pending transactions: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Only mine if there are transactions to process
				if len(pendingTxs.Transactions) == 0 {
					log.Println("No pending transactions, waiting...")
					time.Sleep(5 * time.Second)
					continue
				}

				log.Printf("Found %d pending transactions, starting mining...", len(pendingTxs.Transactions))

				// Create mining request with pending transactions
				req := &pb.MineBlockRequest{
					MinerAddress: minerAddress,
					Transactions: pendingTxs.Transactions,
				}

				// Try to mine a block
				resp, err := c.client.MineBlock(ctx, req)
				if err != nil {
					log.Printf("Mining failed: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				if resp.Success {
					log.Printf("Successfully mined block: %s", resp.BlockHash)
				} else {
					log.Printf("Mining failed: %s", resp.ErrorMessage)
				}

				// Small delay between mining attempts
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Received interrupt signal, stopping miner...")
	cancel()
	time.Sleep(1 * time.Second) // Give a moment for cleanup
}
