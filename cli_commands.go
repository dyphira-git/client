package main

import (
	"fmt"
	"log"

	"dyp_chain/blockchain"
)

func (cli *CLI) createBlockchain(address string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer bc.DB.Close()

	fmt.Println("Done!")
}

func (cli *CLI) createWallet() {
	wallets, _ := blockchain.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) getBalance(address string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := blockchain.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range UTXOs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				balance += out.Value
			}
		}
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) listAddresses() {
	wallets, err := blockchain.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %t\n\n", pow.Validate())

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	tx := blockchain.NewUTXOTransaction(from, to, amount, bc)
	bc.AddTransaction(tx) // Add to mempool instead of directly creating a block
	fmt.Println("Success! Transaction added to mempool.")
}
