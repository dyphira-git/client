package main

import (
	"fmt"
	"log"

	"dyp_chain/blockchain"

	"github.com/ethereum/go-ethereum/common"
)

func (cli *CLI) createBlockchain(address string) {
	if !common.IsHexAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer bc.DB.Close()

	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	if !common.IsHexAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	balance := bc.GetBalance(address)
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %t\n\n", pow.Validate())

		for _, tx := range block.Transactions {
			fmt.Printf("Transaction %x:\n", tx.ID)
			fmt.Printf("  From:   %s\n", tx.From)
			fmt.Printf("  To:     %s\n", tx.To)
			fmt.Printf("  Amount: %d\n\n", tx.Amount)
		}
		fmt.Printf("\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(privateKey, from, to string, amount, fee float32) {
	if !common.IsHexAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !common.IsHexAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	tx := blockchain.NewUTXOTransaction(privateKey, from, to, amount, fee, bc)
	bc.AddTransaction(tx) // Add to mempool instead of directly creating a block
	fmt.Println("Success! Transaction added to mempool.")
}
