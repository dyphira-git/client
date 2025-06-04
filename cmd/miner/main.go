package main

import (
	"flag"
	"log"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Mining server address")
	minerAddr := flag.String("miner", "", "Miner's wallet address")
	flag.Parse()

	if *minerAddr == "" {
		log.Fatal("Please provide a miner wallet address using -miner flag")
	}

	client, err := NewMiningClient(*serverAddr)
	if err != nil {
		log.Fatalf("Failed to create mining client: %v", err)
	}
	defer client.Close()

	log.Printf("Starting mining operations for address: %s", *minerAddr)
	client.StartMining(*minerAddr)
}
