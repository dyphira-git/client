package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func NewServer(port string) *Server {
	return &Server{port: port}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	// Create a new wallet
	mux.HandleFunc("/wallet", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			wallets, _ := NewWallets()
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
		wallet, err := NewWalletFromPrivateKey(req.PrivateKey)
		if err != nil {
			http.Error(w, "Invalid private key: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get the address
		address := string(wallet.GetAddress())

		// Save the wallet
		wallets, _ := NewWallets()
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

		if !ValidateAddress(address) {
			http.Error(w, "Invalid address", http.StatusBadRequest)
			return
		}

		bc := NewBlockchain()
		defer bc.db.Close()

		pubKeyHash := Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := bc.FindUnspentTransactions(pubKeyHash)

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
	mux.HandleFunc("/wallet/addresses", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		wallets, err := NewWallets()
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

			if !ValidateAddress(req.Address) {
				http.Error(w, "Invalid address", http.StatusBadRequest)
				return
			}

			bc := CreateBlockchain(req.Address)
			defer bc.db.Close()

			json.NewEncoder(w).Encode(map[string]string{"message": "Blockchain created successfully"})

		case http.MethodGet:
			// View blockchain
			bc := NewBlockchain()
			defer bc.db.Close()

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

		if !ValidateAddress(req.ToAddress) {
			http.Error(w, "Invalid ToAddress", http.StatusBadRequest)
			return
		}

		// Create wallet from private key
		wallet, err := NewWalletFromPrivateKey(req.PrivateKey)
		if err != nil {
			http.Error(w, "Invalid private key: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get sender's address
		fromAddress := string(wallet.GetAddress())

		bc := NewBlockchain()
		defer bc.db.Close()

		// Get the genesis miner address (the one who created the blockchain)
		minerAddress := bc.GetGenesisAddress()
		if minerAddress == "" {
			http.Error(w, "Could not determine miner address", http.StatusInternalServerError)
			return
		}

		// Create the main transaction
		tx := NewUTXOTransaction(fromAddress, req.ToAddress, req.Amount, bc)

		// Create a mining reward transaction for the genesis miner
		minerReward := NewCoinbaseTX(minerAddress, "Mining reward")

		// Add both transactions to the block
		bc.AddBlock([]*Transaction{tx, minerReward})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Transaction completed successfully",
			"details": map[string]interface{}{
				"transaction": map[string]interface{}{
					"from":   fromAddress,
					"to":     req.ToAddress,
					"amount": req.Amount,
				},
				"mining_reward": map[string]interface{}{
					"amount": 100,
					"to":     minerAddress,
				},
			},
		})
	}))

	log.Printf("Server starting on port %s\n", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, mux))
}
