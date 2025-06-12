package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const MINING_REWARD = float32(50)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID        []byte
	Vin       []TXInput
	Vout      []TXOutput
	From      string
	To        string
	Amount    float32
	Fee       float32
	Signature []byte
}

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(address string) bool {
	// Convert the input's public key to an address
	pubKey, err := crypto.UnmarshalPubkey(in.PubKey)
	if err != nil {
		return false
	}
	inputAddress := crypto.PubkeyToAddress(*pubKey)
	return strings.EqualFold(inputAddress.Hex(), address)
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value   float32
	Address string // address
}

// NewCoinbaseTx creates a new coinbase transaction
func NewCoinbaseTx(to, data string, isGenesis bool, totalFees float32) *Transaction {
	if !common.IsHexAddress(to) {
		log.Panic("Invalid miner address")
	}

	// Mining reward is always 50 DYP plus transaction fees
	reward := MINING_REWARD + totalFees

	txin := TXInput{[]byte{}, -1, nil, nil}
	txout := NewTXOutput(reward, to)
	tx := Transaction{
		ID:        []byte{},
		Vin:       []TXInput{txin},
		Vout:      []TXOutput{*txout},
		From:      "coinbase",
		To:        to,
		Amount:    reward,
		Fee:       0,
		Signature: []byte(data),
	}
	tx.ID = tx.Hash()
	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(privateKeyHex, from, to string, amount, fee float32, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// Validate addresses
	if !common.IsHexAddress(from) || !common.IsHexAddress(to) {
		log.Panic("Invalid address format")
	}

	// Create wallet from private key
	wallet, err := NewWalletFromPrivateKey(privateKeyHex)
	if err != nil {
		log.Panic(err)
	}

	// Total amount needed is amount + fee
	totalNeeded := amount + fee
	acc, validOutputs := bc.FindSpendableOutputs(from, totalNeeded)

	if acc < totalNeeded {
		log.Panic("ERROR: Not enough funds to cover amount and fee")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			// Get the public key bytes
			pubKeyBytes := crypto.FromECDSAPub(&wallet.PrivateKey.PublicKey)
			input := TXInput{txID, out, nil, pubKeyBytes}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *NewTXOutput(amount, to)) // Payment to recipient
	if acc > totalNeeded {
		outputs = append(outputs, *NewTXOutput(acc-totalNeeded, from)) // Change back to sender
	}

	tx := Transaction{
		ID:        nil,
		Vin:       inputs,
		Vout:      outputs,
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Signature: nil,
	}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}
	txCopy.Signature = nil

	data := struct {
		Vin    []TXInput
		Vout   []TXOutput
		From   string
		To     string
		Amount float32
	}{
		Vin:    txCopy.Vin,
		Vout:   txCopy.Vout,
		From:   txCopy.From,
		To:     txCopy.To,
		Amount: txCopy.Amount,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	hash = crypto.Keccak256Hash(buf.Bytes())
	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = []byte(prevTx.Vout[vin.Vout].Address)
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		dataHash := crypto.Keccak256Hash(txCopy.ID)
		signature, err := crypto.Sign(dataHash.Bytes(), privKey)
		if err != nil {
			log.Panic(err)
		}

		tx.Vin[inID].Signature = signature
	}
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = []byte(prevTx.Vout[vin.Vout].Address)
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		dataHash := crypto.Keccak256Hash(txCopy.ID)

		// Recover the public key from the signature
		pubKey, err := crypto.Ecrecover(dataHash.Bytes(), vin.Signature)
		if err != nil {
			return false
		}

		// Verify the signature
		sigPublicKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
		if err != nil {
			return false
		}

		// Get the address from the public key
		recoveredAddr := crypto.PubkeyToAddress(*sigPublicKeyECDSA)
		expectedAddr := common.HexToAddress(string(prevTx.Vout[vin.Vout].Address))

		if recoveredAddr != expectedAddr {
			return false
		}
	}

	return true
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.Address})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, tx.From, tx.To, tx.Amount, tx.Fee, nil}

	return txCopy
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// NewTXOutput creates a new TXOutput
func NewTXOutput(value float32, address string) *TXOutput {
	if !common.IsHexAddress(address) {
		log.Panic("Invalid address format")
	}
	return &TXOutput{value, address}
}
