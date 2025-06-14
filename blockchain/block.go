package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prevHash"`
	Hash         string        `json:"hash"`
	Nonce        int64         `json:"nonce"`
	MerkleRoot   string        `json:"merkleRoot"`
	MerkleTree   *MerkleTree   `json:"-"`
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Fee    float64 `json:"fee"`
	Hash   string  `json:"hash"`
}

// NewBlock creates a new block with Merkle tree integration
func NewBlock(index int64, transactions []Transaction, prevHash string) *Block {
	merkleTree := NewMerkleTree(transactions)
	merkleRoot := ""
	if merkleTree.Root != nil {
		merkleRoot = merkleTree.GetMerkleRoot()
	}

	return &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
		Hash:         "",
		MerkleRoot:   merkleRoot,
		MerkleTree:   merkleTree,
	}
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount, fee float64) *Transaction {
	tx := &Transaction{
		From:   from,
		To:     to,
		Amount: amount,
		Fee:    fee,
	}
	tx.Hash = tx.calculateHash()
	return tx
}

// calculateHash calculates the hash of the block (now includes Merkle root)
func (b *Block) calculateHash() string {
	data := struct {
		Index      int64
		Timestamp  int64
		MerkleRoot string
		PrevHash   string
		Nonce      int64
	}{
		Index:      b.Index,
		Timestamp:  b.Timestamp,
		MerkleRoot: b.MerkleRoot,
		PrevHash:   b.PrevHash,
		Nonce:      b.Nonce,
	}
	blockBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(blockBytes)
	return hex.EncodeToString(hash[:])
}

// calculateHash calculates the hash of the transaction
func (tx *Transaction) calculateHash() string {
	data := struct {
		From   string
		To     string
		Amount float64
		Fee    float64
	}{
		From:   tx.From,
		To:     tx.To,
		Amount: tx.Amount,
		Fee:    tx.Fee,
	}
	txBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(txBytes)
	return hex.EncodeToString(hash[:])
}

// MineBlock mines the block with a given difficulty
func (b *Block) MineBlock(difficulty int) {
	target := make([]byte, difficulty)
	for i := 0; i < difficulty; i++ {
		target[i] = '0'
	}
	targetStr := string(target)

	for {
		b.Nonce++
		b.Hash = b.calculateHash()
		if b.Hash[:difficulty] == targetStr {
			break
		}
	}
}

// ValidateTransactions validates all transactions in the block using Merkle tree
func (b *Block) ValidateTransactions() bool {
	if b.MerkleTree == nil {
		b.MerkleTree = NewMerkleTree(b.Transactions)
		if b.MerkleTree.Root != nil {
			b.MerkleRoot = b.MerkleTree.GetMerkleRoot()
		}
	}

	calculatedRoot := ""
	if b.MerkleTree.Root != nil {
		calculatedRoot = b.MerkleTree.GetMerkleRoot()
	}

	return b.MerkleRoot == calculatedRoot
}

// GenerateTransactionProof generates a Merkle proof for a specific transaction
func (b *Block) GenerateTransactionProof(txHash string) (*MerkleProof, error) {
	if b.MerkleTree == nil {
		b.MerkleTree = NewMerkleTree(b.Transactions)
	}
	return b.MerkleTree.GenerateProof(txHash)
}

// VerifyTransactionProof verifies that a transaction exists in this block
func (b *Block) VerifyTransactionProof(proof *MerkleProof) bool {
	return VerifyProof(proof, b.MerkleRoot)
}
