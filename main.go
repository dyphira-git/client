package main

import (
	"dyp_chain/api"
	"dyp_chain/blockchain"
	pb "dyp_chain/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	// Initialize blockchain
	bc := blockchain.NewBlockchain()

	// Start HTTP server in a goroutine
	go func() {
		server := api.NewServer("8080", bc)
		server.Start()
	}()

	// Create gRPC server
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
