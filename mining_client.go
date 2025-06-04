package main

import (
	"context"
	pb "dyp_chain/proto"
	"fmt"
	"log"
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
	for {
		status, err := c.client.GetBlockchainStatus(context.Background(), &pb.BlockchainStatusRequest{})
		if err != nil {
			log.Printf("Failed to get blockchain status: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("Blockchain status: %+v", status)

		// Create mining request
		req := &pb.MineBlockRequest{
			MinerAddress: minerAddress,
			Transactions: []*pb.Transaction{}, // Empty transactions for now
		}

		// Try to mine a block
		resp, err := c.client.MineBlock(context.Background(), req)
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
