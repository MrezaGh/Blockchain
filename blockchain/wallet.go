package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
)

// Wallet represents a wallet in the blockchain
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// NewWallet creates a new wallet
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := &privateKey.PublicKey
	address := generateAddress(publicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

// generateAddress generates a wallet address from the public key
func generateAddress(publicKey *ecdsa.PublicKey) string {
	// Concatenate X and Y coordinates of the public key
	keyBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)

	// Hash the public key
	hash := sha256.Sum256(keyBytes)

	// Return the hex-encoded hash as the address
	return hex.EncodeToString(hash[:])
}

// SignTransaction signs a transaction with the private key
func (w *Wallet) SignTransaction(tx Transaction) (string, error) {
	// Convert transaction to bytes
	txBytes := []byte(tx.From + tx.To + strconv.FormatFloat(tx.Amount, 'f', -1, 64))

	// Hash the transaction
	hash := sha256.Sum256(txBytes)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return "", err
	}

	// Combine r and s into a single signature
	signature := append(r.Bytes(), s.Bytes()...)

	return hex.EncodeToString(signature), nil
}

// VerifyTransaction verifies a transaction signature
func (w *Wallet) VerifyTransaction(tx Transaction, signature string) bool {
	// Convert transaction to bytes
	txBytes := []byte(tx.From + tx.To + strconv.FormatFloat(tx.Amount, 'f', -1, 64))

	// Hash the transaction
	hash := sha256.Sum256(txBytes)

	// Decode the signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// Split signature into r and s
	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])

	// Verify the signature
	return ecdsa.Verify(w.PublicKey, hash[:], r, s)
}
