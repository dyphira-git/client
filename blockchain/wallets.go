package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallet.dat"

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// SerializedWallet is used for serialization
type SerializedWallet struct {
	PrivateKeyD     []byte
	PrivateKeyX     []byte
	PrivateKeyY     []byte
	PrivateKeyCurve string
	PublicKey       []byte
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var serializedWallets map[string]*SerializedWallet
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&serializedWallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = make(map[string]*Wallet)
	for addr, serializedWallet := range serializedWallets {
		wallet := &Wallet{}
		wallet.PublicKey = serializedWallet.PublicKey

		curve := elliptic.P256()
		wallet.PrivateKey.Curve = curve
		wallet.PrivateKey.D = new(big.Int).SetBytes(serializedWallet.PrivateKeyD)
		wallet.PrivateKey.X = new(big.Int).SetBytes(serializedWallet.PrivateKeyX)
		wallet.PrivateKey.Y = new(big.Int).SetBytes(serializedWallet.PrivateKeyY)

		ws.Wallets[addr] = wallet
	}

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	serializedWallets := make(map[string]*SerializedWallet)

	for addr, wallet := range ws.Wallets {
		serializedWallet := &SerializedWallet{
			PrivateKeyD:     wallet.PrivateKey.D.Bytes(),
			PrivateKeyX:     wallet.PrivateKey.X.Bytes(),
			PrivateKeyY:     wallet.PrivateKey.Y.Bytes(),
			PrivateKeyCurve: "P256",
			PublicKey:       wallet.PublicKey,
		}
		serializedWallets[addr] = serializedWallet
	}

	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(serializedWallets)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
