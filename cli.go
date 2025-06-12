package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendPrivateKey := sendCmd.String("privateKey", "", "The private key of the sender")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Float64("amount", 0, "Amount to send")
	sendFee := sendCmd.Float64("fee", 0, "Fee to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		if !common.IsHexAddress(*createBlockchainAddress) {
			log.Panic("ERROR: Invalid address format")
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		if !common.IsHexAddress(*getBalanceAddress) {
			log.Panic("ERROR: Invalid Ethereum address format")
		}
		cli.getBalance(*getBalanceAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendPrivateKey == "" || *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		if !common.IsHexAddress(*sendFrom) {
			log.Panic("ERROR: Invalid source address format")
		}
		if !common.IsHexAddress(*sendTo) {
			log.Panic("ERROR: Invalid destination address format")
		}
		cli.send(*sendPrivateKey, *sendFrom, *sendTo, float32(*sendAmount), float32(*sendFee))
	}
}
