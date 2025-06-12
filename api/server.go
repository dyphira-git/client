package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"dyp_chain/blockchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

// Server represents the HTTP server for the blockchain API
type Server struct {
	port    string
	bc      *blockchain.Blockchain
	limiter limiter.Store
}

// Response types
type (
	CreateWalletResponse struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
	}

	ErrorResponse struct {
		Error string `json:"error"`
	}

	BalanceResponse struct {
		Address string  `json:"address"`
		Balance float32 `json:"balance"`
	}

	SendRequest struct {
		PrivateKey  string  `json:"private_key"`
		FromAddress string  `json:"from"`
		ToAddress   string  `json:"to"`
		Amount      float32 `json:"amount"`
		Fee         float32 `json:"fee"`
	}

	CreateBlockchainRequest struct {
		Address string `json:"address"`
	}

	ImportWalletRequest struct {
		PrivateKey string `json:"private_key"`
	}

	ImportWalletResponse struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
	}

	TransactionHistoryResponse struct {
		Address string                              `json:"address"`
		History []blockchain.TransactionHistoryItem `json:"history"`
	}

	TransactionResponse struct {
		TxID        string  `json:"txId"`
		From        string  `json:"from"`
		To          string  `json:"to"`
		Amount      float32 `json:"amount"`
		BlockHeight int     `json:"blockHeight"`
		Timestamp   int64   `json:"timestamp"`
		Type        string  `json:"type"`
	}

	AllTransactionsResponse struct {
		Transactions []TransactionResponse `json:"transactions"`
	}

	BlockResponse struct {
		Height        int                   `json:"height"`
		Hash          string                `json:"hash"`
		PrevBlockHash string                `json:"prevBlockHash"`
		Timestamp     int64                 `json:"timestamp"`
		Nonce         int                   `json:"nonce"`
		Transactions  []TransactionResponse `json:"transactions"`
	}

	BlockListResponse struct {
		Blocks []BlockResponse `json:"blocks"`
	}

	TransactionDetailsResponse struct {
		TxID        string  `json:"txId"`
		From        string  `json:"from"`
		To          string  `json:"to"`
		Amount      float32 `json:"amount"`
		Fee         float32 `json:"fee"`
		BlockHeight int     `json:"blockHeight"`
		Timestamp   int64   `json:"timestamp"`
		Type        string  `json:"type"`
		Status      string  `json:"status"` // "confirmed" or "pending"
	}
)

// NewServer creates a new server instance with rate limiting
func NewServer(port string, bc *blockchain.Blockchain) (*Server, error) {
	// Create a new rate limiter that allows 2 requests per second
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   2,           // Number of tokens allowed per interval
		Interval: time.Second, // Interval for token refresh
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %v", err)
	}

	return &Server{
		port:    port,
		bc:      bc,
		limiter: store,
	}, nil
}

// Middleware functions

func (s *Server) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	middleware, err := httplimit.NewMiddleware(s.limiter, httplimit.IPKeyFunc())
	if err != nil {
		log.Printf("Failed to create rate limit middleware: %v", err)
		return next
	}
	return middleware.Handle(next).ServeHTTP
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (s *Server) chainMiddleware(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(middlewares) - 1; i >= 0; i-- {
				last = middlewares[i](last)
			}
			last(w, r)
		}
	}
}

// Handler methods

func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	address := strings.TrimPrefix(r.URL.Path, "/balance/")
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
}

func (s *Server) handleGetAllTransactions(w http.ResponseWriter, r *http.Request) {
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
				BlockHeight: block.Height,
				Timestamp:   block.Timestamp,
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

func (s *Server) handleGetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	address := strings.TrimPrefix(r.URL.Path, "/history/")
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
}

func (s *Server) handleGetAllBlocks(w http.ResponseWriter, r *http.Request) {
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
		txResponses := s.convertTransactions(block)

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

func (s *Server) handleGetSpecificBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	blockHash := strings.TrimPrefix(r.URL.Path, "/block/")
	if blockHash == "" {
		http.Error(w, "Block hash is required", http.StatusBadRequest)
		return
	}

	hash, err := hex.DecodeString(blockHash)
	if err != nil {
		http.Error(w, "Invalid block hash format", http.StatusBadRequest)
		return
	}

	foundBlock := s.findBlock(hash)
	if foundBlock == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	txResponses := s.convertTransactions(foundBlock)
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

func (s *Server) handleSendTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.validateSendRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx := blockchain.NewUTXOTransaction(req.PrivateKey, req.FromAddress, req.ToAddress, req.Amount, req.Fee, s.bc)
	if tx == nil {
		http.Error(w, "Failed to create transaction: insufficient funds", http.StatusBadRequest)
		return
	}

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
}

func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get transaction ID from URL
	txID := strings.TrimPrefix(r.URL.Path, "/transaction/")
	if txID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	// Convert hex string to bytes
	txIDBytes, err := hex.DecodeString(txID)
	if err != nil {
		http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
		return
	}

	// First check mempool for pending transactions
	pendingTxs := s.bc.GetPendingTransactions()
	for _, tx := range pendingTxs {
		if bytes.Equal(tx.ID, txIDBytes) {
			response := TransactionDetailsResponse{
				TxID:   txID,
				From:   tx.From,
				To:     tx.To,
				Amount: tx.Amount,
				Fee:    tx.Fee,
				Type:   "transfer",
				Status: "pending",
			}

			if tx.IsCoinbase() {
				response.Type = "mining_reward"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// If not in mempool, search in blockchain
	tx, err := s.bc.FindTransaction(txIDBytes)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Find the block containing this transaction to get height and timestamp
	var blockHeight int
	var timestamp int64
	bci := s.bc.Iterator()
	for {
		block := bci.Next()
		for _, btx := range block.Transactions {
			if bytes.Equal(btx.ID, txIDBytes) {
				blockHeight = block.Height
				timestamp = block.Timestamp
				break
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	response := TransactionDetailsResponse{
		TxID:        txID,
		From:        tx.From,
		To:          tx.To,
		Amount:      tx.Amount,
		Fee:         tx.Fee,
		BlockHeight: blockHeight,
		Timestamp:   timestamp,
		Type:        "transfer",
		Status:      "confirmed",
	}

	if tx.IsCoinbase() {
		response.Type = "mining_reward"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (s *Server) convertTransactions(block *blockchain.Block) []TransactionResponse {
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
	return txResponses
}

func (s *Server) findBlock(hash []byte) *blockchain.Block {
	bci := s.bc.Iterator()
	for {
		block := bci.Next()
		if bytes.Equal(block.Hash, hash) {
			return block
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

func (s *Server) validateSendRequest(req SendRequest) error {
	if req.PrivateKey == "" || req.ToAddress == "" || req.FromAddress == "" {
		return fmt.Errorf("PrivateKey, FromAddress and ToAddress are required")
	}

	if req.Amount <= 0 {
		return fmt.Errorf("Amount must be greater than 0")
	}

	if req.Fee < 0 {
		return fmt.Errorf("Fee cannot be negative")
	}

	if !common.IsHexAddress(req.ToAddress) {
		return fmt.Errorf("Invalid destination address format")
	}

	return nil
}

// Start initializes and starts the HTTP server
func (s *Server) Start() {
	mux := http.NewServeMux()

	// Apply middleware chain to all handlers
	middleware := s.chainMiddleware(enableCORS, s.rateLimitMiddleware)

	// Register routes
	mux.HandleFunc("/balance/", middleware(s.handleGetBalance))
	mux.HandleFunc("/transactions", middleware(s.handleGetAllTransactions))
	mux.HandleFunc("/history/", middleware(s.handleGetTransactionHistory))
	mux.HandleFunc("/blocks", middleware(s.handleGetAllBlocks))
	mux.HandleFunc("/block/", middleware(s.handleGetSpecificBlock))
	mux.HandleFunc("/transaction", middleware(s.handleSendTransaction))
	mux.HandleFunc("/transaction/", middleware(s.handleGetTransaction))

	log.Printf("Server starting on port %s\n", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, mux))
}
