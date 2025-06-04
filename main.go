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
	loadEnv()

	var bc *blockchain.Blockchain

	if _, err := os.Stat("blockchain.db"); os.IsNotExist(err) {
		bc = blockchain.CreateBlockchain(os.Getenv("GENESIS_ADDRESS"))
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

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
