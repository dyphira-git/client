package main

import (
	"dyp_chain/api"
	"dyp_chain/blockchain"
	pb "dyp_chain/proto"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	// Try to load .env file but don't fail if it doesn't exist
	_ = godotenv.Load()

	// Check if GENESIS_ADDRESS is set
	genesisAddr := os.Getenv("GENESIS_ADDRESS")
	if genesisAddr == "" {
		log.Fatal("GENESIS_ADDRESS environment variable is required")
	}

	var bc *blockchain.Blockchain

	if _, err := os.Stat("blockchain.db"); os.IsNotExist(err) {
		bc = blockchain.CreateBlockchain(genesisAddr)
	} else {
		bc = blockchain.NewBlockchain()
	}

	go func() {
		server := api.NewServer("8080", bc)
		server.Start()
	}()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMiningServiceServer(s, newMiningServer(bc))

	log.Printf("Mining server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
