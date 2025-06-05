package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet stores private and public keys for Ethereum-style addresses
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    common.Address
}

// NewWalletFromPrivateKey creates a wallet from an existing private key hex string
func NewWalletFromPrivateKey(privateKeyHex string) (*Wallet, error) {
	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Parse the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	// Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	// Get public key bytes
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	// Get Ethereum address
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKeyBytes,
		Address:    address,
	}

	return wallet, nil
}

// GetAddress returns the wallet's Ethereum address as a hex string with 0x prefix
func (w *Wallet) GetAddress() string {
	return w.Address.Hex()
}

// GetPrivateKey returns the wallet's private key as a hex string with 0x prefix
func (w *Wallet) GetPrivateKey() string {
	return "0x" + hex.EncodeToString(crypto.FromECDSA(w.PrivateKey))
}

// ValidateAddress checks if the given address is a valid Ethereum address
func ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

// Sign signs the given data with the wallet's private key
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	// Hash the data first
	hash := crypto.Keccak256Hash(data)

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}

	return signature, nil
}

// VerifySignature verifies if the signature was signed by the given address
func VerifySignature(data []byte, signature []byte, address common.Address) bool {
	// Hash the data
	hash := crypto.Keccak256Hash(data)

	// Get public key from signature
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return false
	}

	// Convert to address
	recoveredAddress := common.BytesToAddress(crypto.Keccak256(sigPublicKey[1:])[12:])

	return address == recoveredAddress
}
