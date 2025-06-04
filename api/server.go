package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"dyp_chain/blockchain"
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
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type SendRequest struct {
	PrivateKey string `json:"private_key"`
	ToAddress  string `json:"to_address"`
	Amount     int    `json:"amount"`
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

func NewServer(port string, bc *blockchain.Blockchain) *Server {
	return &Server{port: port, bc: bc}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	// Create a new wallet
	mux.HandleFunc("/wallet", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			wallets, _ := blockchain.NewWallets()
			address := wallets.CreateWallet()
			wallet := wallets.GetWallet(address)

			// Convert private key to hex string
			privateKeyBytes := wallet.PrivateKey.D.Bytes()
			privateKeyHex := hex.EncodeToString(privateKeyBytes)

			wallets.SaveToFile()

			json.NewEncoder(w).Encode(CreateWalletResponse{
				Address:    address,
				PrivateKey: privateKeyHex,
			})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Import wallet
	mux.HandleFunc("/wallet/import", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ImportWalletRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.PrivateKey == "" {
			http.Error(w, "PrivateKey is required", http.StatusBadRequest)
			return
		}

		// Create wallet from private key
		wallet, err := blockchain.NewWalletFromPrivateKey(req.PrivateKey)
		if err != nil {
			http.Error(w, "Invalid private key: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get the address
		address := string(wallet.GetAddress())

		// Save the wallet
		wallets, _ := blockchain.NewWallets()
		wallets.Wallets[address] = wallet
		wallets.SaveToFile()

		json.NewEncoder(w).Encode(ImportWalletResponse{
			Address:    address,
			PrivateKey: req.PrivateKey,
		})
	}))

	// Get balance
	mux.HandleFunc("/wallet/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Skip "/wallet/" prefix and "/balance" suffix to get the address
		path := r.URL.Path[len("/wallet/"):]
		if !strings.HasSuffix(path, "/balance") {
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			return
		}

		address := path[:len(path)-len("/balance")]
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		if !blockchain.ValidateAddress(address) {
			http.Error(w, "Invalid address", http.StatusBadRequest)
			return
		}

		pubKeyHash := blockchain.Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := s.bc.FindUnspentTransactions(pubKeyHash)

		balance := 0
		for _, tx := range UTXOs {
			for _, out := range tx.Vout {
				if out.IsLockedWithKey(pubKeyHash) {
					balance += out.Value
				}
			}
		}

		json.NewEncoder(w).Encode(BalanceResponse{
			Address: address,
			Balance: balance,
		})
	}))

	// List all addresses
	mux.HandleFunc("/wallets", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		wallets, err := blockchain.NewWallets()
		if err != nil {
			http.Error(w, "Failed to get wallets", http.StatusInternalServerError)
			return
		}

		addresses := wallets.GetAddresses()
		json.NewEncoder(w).Encode(map[string][]string{"addresses": addresses})
	}))

	// Blockchain operations (create and view)
	mux.HandleFunc("/blockchain", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Create blockchain
			var req CreateBlockchainRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if !blockchain.ValidateAddress(req.Address) {
				http.Error(w, "Invalid address", http.StatusBadRequest)
				return
			}

			bc := blockchain.CreateBlockchain(req.Address)
			defer bc.DB.Close()

			json.NewEncoder(w).Encode(map[string]string{"message": "Blockchain created successfully"})

		case http.MethodGet:
			// View blockchain
			bc := blockchain.NewBlockchain()
			defer bc.DB.Close()

			bci := bc.Iterator()

			var blocks []map[string]interface{}
			for {
				block := bci.Next()

				var transactions []map[string]interface{}
				for _, tx := range block.Transactions {
					transactions = append(transactions, map[string]interface{}{
						"id":   string(tx.ID),
						"vin":  tx.Vin,
						"vout": tx.Vout,
					})
				}

				blocks = append(blocks, map[string]interface{}{
					"hash":         string(block.Hash),
					"transactions": transactions,
					"timestamp":    block.Timestamp,
				})

				if len(block.PrevBlockHash) == 0 {
					break
				}
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"blocks": blocks,
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Send coins
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
		if req.PrivateKey == "" || req.ToAddress == "" {
			http.Error(w, "PrivateKey and ToAddress are required", http.StatusBadRequest)
			return
		}

		if req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}

		if !blockchain.ValidateAddress(req.ToAddress) {
			http.Error(w, "Invalid ToAddress", http.StatusBadRequest)
			return
		}

		// Create wallet from private key
		wallet, err := blockchain.NewWalletFromPrivateKey(req.PrivateKey)
		if err != nil {
			http.Error(w, "Invalid private key: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get sender's address
		fromAddress := string(wallet.GetAddress())

		// Create the transaction
		tx := blockchain.NewUTXOTransaction(fromAddress, req.ToAddress, req.Amount, s.bc)
		if tx == nil {
			http.Error(w, "Failed to create transaction: insufficient funds", http.StatusBadRequest)
			return
		}

		// Add transaction to mempool
		s.bc.AddTransaction(tx)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Transaction added to mempool",
			"details": map[string]interface{}{
				"transaction": map[string]interface{}{
					"from":   fromAddress,
					"to":     req.ToAddress,
					"amount": req.Amount,
				},
			},
		})
	}))

	mux.HandleFunc("/transactions", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		transactions := s.bc.GetAllTransactions()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	}))

	// Transaction history
	mux.HandleFunc("/transaction/history/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract address from URL path
		address := r.URL.Path[len("/transaction/history/"):]
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		history, err := s.bc.GetTransactionHistory(address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := TransactionHistoryResponse{
			Address: address,
			History: history,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	log.Printf("Server starting on port %s\n", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, mux))
}
