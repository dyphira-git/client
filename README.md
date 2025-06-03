# Blockchain API Documentation

This is a simple blockchain implementation with a REST API interface. The blockchain features a single miner (genesis address) that receives all mining rewards.

## Getting Started

1. Install Go 1.21 or later
2. Install dependencies:
```bash
go mod download
```
3. Run the server:
```bash
go run .
```
The server will start on port 8080.

## API Endpoints

### 1. Create a New Wallet
Creates a new wallet and returns its address and private key.

```http
POST /wallet
```

Response:
```json
{
    "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
    "private_key": "e9873d79c6d87dc0fb6a5778633389f4453213303da61f20bd67fc233aa33262"
}
```

### 2. Import Wallet
Import an existing wallet using its private key.

```http
POST /wallet/import
```

Request body:
```json
{
    "private_key": "e9873d79c6d87dc0fb6a5778633389f4453213303da61f20bd67fc233aa33262"
}
```

Response:
```json
{
    "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
    "private_key": "e9873d79c6d87dc0fb6a5778633389f4453213303da61f20bd67fc233aa33262"
}
```

### 3. Get Wallet Balance
Get the balance of a specific wallet address.

```http
GET /wallet/{address}/balance
```

Response:
```json
{
    "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
    "balance": 100
}
```

### 4. List All Wallet Addresses
Get a list of all wallet addresses in the system.

```http
GET /wallet/addresses
```

Response:
```json
{
    "addresses": [
        "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
        "1QAndHBtr6NZyHU4xgJu9AaYELuEcdL1Hp"
    ]
}
```

### 5. Create Blockchain
Initialize the blockchain with a genesis block. The provided address becomes the designated miner who receives all mining rewards.

```http
POST /blockchain
```

Request body:
```json
{
    "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
}
```

Response:
```json
{
    "message": "Blockchain created successfully"
}
```

### 6. View Blockchain
Get all blocks in the blockchain.

```http
GET /blockchain
```

Response:
```json
{
    "blocks": [
        {
            "hash": "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
            "transactions": [
                {
                    "id": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
                    "vin": [...],
                    "vout": [...]
                }
            ],
            "timestamp": 1231006505
        }
    ]
}
```

### 7. Send Transaction
Send coins from one address to another. Uses the sender's private key for security.

```http
POST /transaction
```

Request body:
```json
{
    "private_key": "e9873d79c6d87dc0fb6a5778633389f4453213303da61f20bd67fc233aa33262",
    "to_address": "1QAndHBtr6NZyHU4xgJu9AaYELuEcdL1Hp",
    "amount": 10
}
```

Response:
```json
{
    "message": "Transaction completed successfully",
    "details": {
        "transaction": {
            "from": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
            "to": "1QAndHBtr6NZyHU4xgJu9AaYELuEcdL1Hp",
            "amount": 10
        },
        "mining_reward": {
            "amount": 100,
            "to": "GENESIS_MINER_ADDRESS"
        }
    }
}
```

## Example Workflow

1. Create a miner wallet (this will be the genesis miner):
```bash
curl -X POST http://localhost:8080/wallet
```
Save the address and private key.

2. Initialize the blockchain with the miner's address:
```bash
curl -X POST http://localhost:8080/blockchain \
  -H "Content-Type: application/json" \
  -d '{
    "address": "MINER_ADDRESS"
  }'
```

3. Create another wallet for transactions:
```bash
curl -X POST http://localhost:8080/wallet
```
Save the address and private key.

4. Send a transaction:
```bash
curl -X POST http://localhost:8080/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "private_key": "SENDER_PRIVATE_KEY",
    "to_address": "RECIPIENT_ADDRESS",
    "amount": 10
  }'
```

5. Check balances:
```bash
curl http://localhost:8080/wallet/ADDRESS/balance
```

## Important Notes

1. The genesis miner (first address that creates the blockchain) receives all mining rewards (100 coins per block).
2. Private keys should be kept secure and never shared.
3. All transactions require the sender's private key rather than their address for security.
4. The blockchain data is stored in `blockchain.db` and wallet data in `wallet.dat`.
5. Mining rewards are automatically sent to the genesis miner for each transaction.

## Security Considerations

1. Use HTTPS in production
2. Implement rate limiting
3. Add authentication for sensitive endpoints
4. Never share private keys
5. Keep the genesis miner's private key especially secure 