package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"dyp_chain/blockchain"

	"github.com/ethereum/go-ethereum/common"
)

// enableCORS wraps an http.HandlerFunc and adds CORS headers to the response
func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the original handler
		handler(w, r)
	}
}

type Server struct {
	port string
	bc   *blockchain.Blockchain
}

type CreateWalletResponse struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type BalanceResponse struct {
	Address string  `json:"address"`
	Balance float32 `json:"balance"`
}

type SendRequest struct {
	PrivateKey  string  `json:"private_key"`
	FromAddress string  `json:"from"`
	ToAddress   string  `json:"to"`
	Amount      float32 `json:"amount"`
	Fee         float32 `json:"fee"`
}

type CreateBlockchainRequest struct {
	Address string `json:"address"`
}

type ImportWalletRequest struct {
	PrivateKey string `json:"private_key"`
}

type ImportWalletResponse struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
}

// TransactionHistoryResponse represents the response structure for transaction history
type TransactionHistoryResponse struct {
	Address string                              `json:"address"`
	History []blockchain.TransactionHistoryItem `json:"history"`
}

// TransactionResponse represents a single transaction in the response
type TransactionResponse struct {
	TxID        string  `json:"txId"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Amount      float32 `json:"amount"`
	BlockHeight int     `json:"blockHeight"`
	Timestamp   int64   `json:"timestamp"`
	Type        string  `json:"type"`
}

// AllTransactionsResponse represents the response structure for all transactions
type AllTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}

// BlockResponse represents a single block in the response
type BlockResponse struct {
	Height        int                   `json:"height"`
	Hash          string                `json:"hash"`
	PrevBlockHash string                `json:"prevBlockHash"`
	Timestamp     int64                 `json:"timestamp"`
	Nonce         int                   `json:"nonce"`
	Transactions  []TransactionResponse `json:"transactions"`
}

// BlockListResponse represents the response structure for all blocks
type BlockListResponse struct {
	Blocks []BlockResponse `json:"blocks"`
}

func NewServer(port string, bc *blockchain.Blockchain) *Server {
	return &Server{port: port, bc: bc}
}

// GetAllTransactions returns all transactions in the blockchain
func (s *Server) getAllTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bci := s.bc.Iterator()
	response := AllTransactionsResponse{
		Transactions: make([]TransactionResponse, 0),
	}

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txType := "transfer"
			if tx.IsCoinbase() {
				txType = "mining_reward"
			}

			response.Transactions = append(response.Transactions, TransactionResponse{
				TxID:        hex.EncodeToString(tx.ID),
				From:        tx.From,
				To:          tx.To,
				Amount:      tx.Amount,
				BlockHeight: block.Height,    // Include block height
				Timestamp:   block.Timestamp, // Include block timestamp
				Type:        txType,
			})
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllBlocks returns all blocks in the blockchain
func (s *Server) getAllBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bci := s.bc.Iterator()
	response := BlockListResponse{
		Blocks: make([]BlockResponse, 0),
	}

	for {
		block := bci.Next()

		// Convert transactions
		txResponses := make([]TransactionResponse, len(block.Transactions))
		for i, tx := range block.Transactions {
			txType := "transfer"
			if tx.IsCoinbase() {
				txType = "mining_reward"
			}

			txResponses[i] = TransactionResponse{
				TxID:        hex.EncodeToString(tx.ID),
				From:        tx.From,
				To:          tx.To,
				Amount:      tx.Amount,
				BlockHeight: block.Height,
				Timestamp:   block.Timestamp,
				Type:        txType,
			}
		}

		blockResponse := BlockResponse{
			Height:        block.Height,
			Hash:          hex.EncodeToString(block.Hash),
			PrevBlockHash: hex.EncodeToString(block.PrevBlockHash),
			Timestamp:     block.Timestamp,
			Nonce:         block.Nonce,
			Transactions:  txResponses,
		}

		response.Blocks = append(response.Blocks, blockResponse)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSpecificBlock returns a specific block by hash
func (s *Server) getSpecificBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get block hash from URL
	blockHash := r.URL.Path[len("/block/"):]
	if blockHash == "" {
		http.Error(w, "Block hash is required", http.StatusBadRequest)
		return
	}

	// Decode hex string to bytes
	hash, err := hex.DecodeString(blockHash)
	if err != nil {
		http.Error(w, "Invalid block hash format", http.StatusBadRequest)
		return
	}

	// Find the block
	var foundBlock *blockchain.Block
	bci := s.bc.Iterator()
	for {
		block := bci.Next()
		if bytes.Equal(block.Hash, hash) {
			foundBlock = block
			break
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	if foundBlock == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	// Convert transactions
	txResponses := make([]TransactionResponse, len(foundBlock.Transactions))
	for i, tx := range foundBlock.Transactions {
		txType := "transfer"
		if tx.IsCoinbase() {
			txType = "mining_reward"
		}

		txResponses[i] = TransactionResponse{
			TxID:        hex.EncodeToString(tx.ID),
			From:        tx.From,
			To:          tx.To,
			Amount:      tx.Amount,
			BlockHeight: foundBlock.Height,
			Timestamp:   foundBlock.Timestamp,
			Type:        txType,
		}
	}

	blockResponse := BlockResponse{
		Height:        foundBlock.Height,
		Hash:          hex.EncodeToString(foundBlock.Hash),
		PrevBlockHash: hex.EncodeToString(foundBlock.PrevBlockHash),
		Timestamp:     foundBlock.Timestamp,
		Nonce:         foundBlock.Nonce,
		Transactions:  txResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blockResponse)
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	// Get balance
	mux.HandleFunc("/balance/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Skip "/balance/" prefix to get the address
		path := r.URL.Path[len("/balance/"):]

		address := path
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		if !common.IsHexAddress(address) {
			http.Error(w, "Invalid address format", http.StatusBadRequest)
			return
		}

		balance := s.bc.GetBalance(address)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BalanceResponse{
			Address: address,
			Balance: balance,
		})
	}))

	// Get all transactions
	mux.HandleFunc("/transactions", enableCORS(s.getAllTransactions))

	// Get transaction history
	mux.HandleFunc("/history/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Skip "/history/" prefix to get the address
		path := r.URL.Path[len("/history/"):]

		address := path
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		if !common.IsHexAddress(address) {
			http.Error(w, "Invalid address format", http.StatusBadRequest)
			return
		}

		history, err := s.bc.GetTransactionHistory(address)
		if err != nil {
			http.Error(w, "Failed to get transaction history: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TransactionHistoryResponse{
			Address: address,
			History: history,
		})
	}))

	// Get all blocks
	mux.HandleFunc("/blocks", enableCORS(s.getAllBlocks))

	// Get specific block by hash
	mux.HandleFunc("/block/", enableCORS(s.getSpecificBlock))

	// Send transaction
	mux.HandleFunc("/transaction", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate request fields
		if req.PrivateKey == "" || req.ToAddress == "" || req.FromAddress == "" {
			http.Error(w, "PrivateKey, FromAddress and ToAddress are required", http.StatusBadRequest)
			return
		}

		if req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}

		if req.Fee < 0 {
			http.Error(w, "Fee cannot be negative", http.StatusBadRequest)
			return
		}

		if !common.IsHexAddress(req.ToAddress) {
			http.Error(w, "Invalid destination address format", http.StatusBadRequest)
			return
		}

		// Create the transaction
		tx := blockchain.NewUTXOTransaction(req.PrivateKey, req.FromAddress, req.ToAddress, req.Amount, req.Fee, s.bc)
		if tx == nil {
			http.Error(w, "Failed to create transaction: insufficient funds", http.StatusBadRequest)
			return
		}

		// Add transaction to mempool
		s.bc.AddTransaction(tx)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Transaction added to mempool",
			"details": map[string]interface{}{
				"transaction": map[string]interface{}{
					"from":   req.FromAddress,
					"to":     req.ToAddress,
					"amount": req.Amount,
					"fee":    req.Fee,
				},
			},
		})
	}))

	log.Printf("Server starting on port %s\n", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, mux))
}
