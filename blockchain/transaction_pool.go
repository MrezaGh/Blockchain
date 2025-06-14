package blockchain

import (
	"errors"
	"sync"
)

// TransactionPool represents the mempool of pending transactions
type TransactionPool struct {
	transactions map[string]*Transaction
	mu           sync.RWMutex
	maxSize      int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxSize int) *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*Transaction),
		maxSize:      maxSize,
	}
}

// AddTransaction adds a transaction to the pool if it's valid
func (tp *TransactionPool) AddTransaction(tx *Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Check pool size
	if len(tp.transactions) >= tp.maxSize {
		return errors.New("transaction pool is full")
	}

	// Validate transaction
	if err := tp.validateTransaction(tx); err != nil {
		return err
	}

	// Add transaction to pool
	tp.transactions[tx.Hash] = tx
	return nil
}

// GetTransactions returns all transactions in the pool
func (tp *TransactionPool) GetTransactions() []*Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	txs := make([]*Transaction, 0, len(tp.transactions))
	for _, tx := range tp.transactions {
		txs = append(txs, tx)
	}
	return txs
}

// RemoveTransactions removes transactions from the pool
func (tp *TransactionPool) RemoveTransactions(txs []*Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for _, tx := range txs {
		delete(tp.transactions, tx.Hash)
	}
}

// validateTransaction validates a transaction
func (tp *TransactionPool) validateTransaction(tx *Transaction) error {
	// Basic validation
	if tx.From == "" || tx.To == "" {
		return errors.New("invalid transaction: missing from/to address")
	}

	if tx.Amount <= 0 {
		return errors.New("invalid transaction: amount must be positive")
	}

	if tx.Fee < 0 {
		return errors.New("invalid transaction: fee cannot be negative")
	}

	// Check if transaction already exists
	if _, exists := tp.transactions[tx.Hash]; exists {
		return errors.New("transaction already exists in pool")
	}

	return nil
}
