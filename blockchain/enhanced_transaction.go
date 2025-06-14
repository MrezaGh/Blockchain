package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// TransactionType represents different types of transactions
type TransactionType string

const (
	StandardTx TransactionType = "standard"
	MultiSigTx TransactionType = "multisig"
	TimeLockTx TransactionType = "timelock"
	ContractTx TransactionType = "contract"
)

// EnhancedTransaction represents an enhanced transaction with additional features
type EnhancedTransaction struct {
	ID         string                 `json:"id"`
	Type       TransactionType        `json:"type"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Amount     float64                `json:"amount"`
	Fee        float64                `json:"fee"`
	Timestamp  int64                  `json:"timestamp"`
	Hash       string                 `json:"hash"`
	Signatures []TransactionSignature `json:"signatures"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`

	// Multi-signature fields
	RequiredSigs int      `json:"requiredSigs,omitempty"`
	Signers      []string `json:"signers,omitempty"`

	// Time-lock fields
	LockTime     int64 `json:"lockTime,omitempty"`     // Unix timestamp when transaction can be executed
	LockDuration int64 `json:"lockDuration,omitempty"` // Duration in seconds from creation

	// Contract fields
	ContractCode string `json:"contractCode,omitempty"`
	ContractData string `json:"contractData,omitempty"`
}

// TransactionSignature represents a signature with the signer's public key
type TransactionSignature struct {
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
	Signer    string `json:"signer"`
}

// NewStandardTransaction creates a standard transaction
func NewStandardTransaction(from, to string, amount, fee float64, metadata map[string]interface{}) *EnhancedTransaction {
	tx := &EnhancedTransaction{
		Type:       StandardTx,
		From:       from,
		To:         to,
		Amount:     amount,
		Fee:        fee,
		Timestamp:  time.Now().Unix(),
		Metadata:   metadata,
		Signatures: make([]TransactionSignature, 0),
	}
	tx.ID = tx.generateID()
	tx.Hash = tx.calculateHash()
	return tx
}

// NewMultiSigTransaction creates a multi-signature transaction
func NewMultiSigTransaction(from, to string, amount, fee float64, requiredSigs int, signers []string, metadata map[string]interface{}) *EnhancedTransaction {
	tx := &EnhancedTransaction{
		Type:         MultiSigTx,
		From:         from,
		To:           to,
		Amount:       amount,
		Fee:          fee,
		Timestamp:    time.Now().Unix(),
		RequiredSigs: requiredSigs,
		Signers:      signers,
		Metadata:     metadata,
		Signatures:   make([]TransactionSignature, 0),
	}
	tx.ID = tx.generateID()
	tx.Hash = tx.calculateHash()
	return tx
}

// NewTimeLockTransaction creates a time-locked transaction
func NewTimeLockTransaction(from, to string, amount, fee float64, lockTime int64, metadata map[string]interface{}) *EnhancedTransaction {
	tx := &EnhancedTransaction{
		Type:       TimeLockTx,
		From:       from,
		To:         to,
		Amount:     amount,
		Fee:        fee,
		Timestamp:  time.Now().Unix(),
		LockTime:   lockTime,
		Metadata:   metadata,
		Signatures: make([]TransactionSignature, 0),
	}
	tx.ID = tx.generateID()
	tx.Hash = tx.calculateHash()
	return tx
}

// generateID generates a unique transaction ID
func (tx *EnhancedTransaction) generateID() string {
	data := struct {
		Type      TransactionType
		From      string
		To        string
		Amount    float64
		Timestamp int64
	}{
		Type:      tx.Type,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Timestamp: tx.Timestamp,
	}

	bytes, _ := json.Marshal(data)
	return calculateHashFromBytes(bytes)
}

// calculateHash calculates the transaction hash
func (tx *EnhancedTransaction) calculateHash() string {
	data := struct {
		ID           string
		Type         TransactionType
		From         string
		To           string
		Amount       float64
		Fee          float64
		Timestamp    int64
		RequiredSigs int
		Signers      []string
		LockTime     int64
		Metadata     map[string]interface{}
	}{
		ID:           tx.ID,
		Type:         tx.Type,
		From:         tx.From,
		To:           tx.To,
		Amount:       tx.Amount,
		Fee:          tx.Fee,
		Timestamp:    tx.Timestamp,
		RequiredSigs: tx.RequiredSigs,
		Signers:      tx.Signers,
		LockTime:     tx.LockTime,
		Metadata:     tx.Metadata,
	}

	bytes, _ := json.Marshal(data)
	return calculateHashFromBytes(bytes)
}

// calculateHashFromBytes calculates hash from byte slice
func calculateHashFromBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AddSignature adds a signature to the transaction
func (tx *EnhancedTransaction) AddSignature(signature TransactionSignature) error {
	// Verify the signature is valid for this transaction
	if !tx.verifySignature(signature) {
		return errors.New("invalid signature")
	}

	// Check if signer is already signed (prevent duplicate signatures)
	for _, sig := range tx.Signatures {
		if sig.Signer == signature.Signer {
			return errors.New("transaction already signed by this signer")
		}
	}

	// For multi-sig transactions, verify signer is authorized
	if tx.Type == MultiSigTx {
		authorized := false
		for _, signer := range tx.Signers {
			if signer == signature.Signer {
				authorized = true
				break
			}
		}
		if !authorized {
			return errors.New("signer not authorized for this multi-sig transaction")
		}
	}

	tx.Signatures = append(tx.Signatures, signature)
	return nil
}

// IsFullySigned checks if the transaction has sufficient signatures
func (tx *EnhancedTransaction) IsFullySigned() bool {
	switch tx.Type {
	case StandardTx:
		return len(tx.Signatures) >= 1
	case MultiSigTx:
		return len(tx.Signatures) >= tx.RequiredSigs
	case TimeLockTx:
		return len(tx.Signatures) >= 1
	case ContractTx:
		return len(tx.Signatures) >= 1
	default:
		return false
	}
}

// IsExecutable checks if the transaction can be executed (considers time locks)
func (tx *EnhancedTransaction) IsExecutable() bool {
	if !tx.IsFullySigned() {
		return false
	}

	// Check time lock conditions
	if tx.Type == TimeLockTx && tx.LockTime > 0 {
		return time.Now().Unix() >= tx.LockTime
	}

	return true
}

// verifySignature verifies a signature against the transaction
func (tx *EnhancedTransaction) verifySignature(sig TransactionSignature) bool {
	// This is a simplified verification - in a real implementation,
	// you would use the actual ECDSA verification with the public key
	return len(sig.Signature) > 0 && len(sig.PublicKey) > 0 && len(sig.Signer) > 0
}

// GetMetadata retrieves metadata value by key
func (tx *EnhancedTransaction) GetMetadata(key string) (interface{}, bool) {
	if tx.Metadata == nil {
		return nil, false
	}
	value, exists := tx.Metadata[key]
	return value, exists
}

// SetMetadata sets a metadata key-value pair
func (tx *EnhancedTransaction) SetMetadata(key string, value interface{}) {
	if tx.Metadata == nil {
		tx.Metadata = make(map[string]interface{})
	}
	tx.Metadata[key] = value
	// Recalculate hash after metadata change
	tx.Hash = tx.calculateHash()
}

// ToStandardTransaction converts enhanced transaction to standard transaction for backward compatibility
func (tx *EnhancedTransaction) ToStandardTransaction() Transaction {
	return Transaction{
		From:   tx.From,
		To:     tx.To,
		Amount: tx.Amount,
		Fee:    tx.Fee,
		Hash:   tx.Hash,
	}
}

// SignTransactionEnhanced signs an enhanced transaction with a wallet
func (w *Wallet) SignTransactionEnhanced(tx *EnhancedTransaction) (*TransactionSignature, error) {
	// Sign the transaction hash
	signature, err := w.SignTransaction(tx.ToStandardTransaction())
	if err != nil {
		return nil, err
	}

	// Create transaction signature
	txSig := &TransactionSignature{
		PublicKey: publicKeyToString(w.PublicKey),
		Signature: signature,
		Signer:    w.Address,
	}

	return txSig, nil
}

// Helper function to convert public key to string (simplified)
func publicKeyToString(pubKey *ecdsa.PublicKey) string {
	return pubKey.X.String() + ":" + pubKey.Y.String()
}
