# Simple Blockchain Implementation in Go

This is a simple blockchain implementation in Go, featuring:

- Basic blockchain structure with proof-of-work
- Transactions with public-key cryptography
- CLI interface for interacting with the blockchain
- Wallet management
- UTXO (Unspent Transaction Output) model

## Prerequisites

- Go 1.21 or later
- `github.com/boltdb/bolt` package
- `golang.org/x/crypto/ripemd160` package

## Installation

1. Clone the repository
2. Install dependencies:
```bash
go mod download
```

## Usage

The application provides the following commands:

1. Create a new wallet:
```bash
go run . createwallet
```

2. Create a blockchain with initial balance:
```bash
go run . createblockchain -address ADDRESS
```

3. Get balance for an address:
```bash
go run . getbalance -address ADDRESS
```

4. Send coins from one address to another:
```bash
go run . send -from FROM_ADDRESS -to TO_ADDRESS -amount AMOUNT
```

5. List all wallet addresses:
```bash
go run . listaddresses
```

6. Print the blockchain:
```bash
go run . printchain
```

## Example Usage

1. First, create two wallets:
```bash
go run . createwallet
go run . createwallet
```

2. Create a blockchain and send the genesis reward to the first address:
```bash
go run . createblockchain -address FIRST_ADDRESS
```

3. Check the balance of the first address:
```bash
go run . getbalance -address FIRST_ADDRESS
```

4. Send some coins to the second address:
```bash
go run . send -from FIRST_ADDRESS -to SECOND_ADDRESS -amount 10
```

5. Check the balances:
```bash
go run . getbalance -address FIRST_ADDRESS
go run . getbalance -address SECOND_ADDRESS
```

## Note

This is a simplified implementation for educational purposes. It implements the core concepts of a blockchain but should not be used in production. 