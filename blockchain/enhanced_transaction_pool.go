package blockchain

import (
	"errors"
	"sync"
	"time"
)

// EnhancedTransactionPool manages enhanced transactions with additional validation
type EnhancedTransactionPool struct {
	standardTxs map[string]*Transaction         // Standard transactions
	enhancedTxs map[string]*EnhancedTransaction // Enhanced transactions
	mu          sync.RWMutex
	maxSize     int
}

// NewEnhancedTransactionPool creates a new enhanced transaction pool
func NewEnhancedTransactionPool(maxSize int) *EnhancedTransactionPool {
	return &EnhancedTransactionPool{
		standardTxs: make(map[string]*Transaction),
		enhancedTxs: make(map[string]*EnhancedTransaction),
		maxSize:     maxSize,
	}
}

// AddStandardTransaction adds a standard transaction to the pool
func (etp *EnhancedTransactionPool) AddStandardTransaction(tx *Transaction) error {
	etp.mu.Lock()
	defer etp.mu.Unlock()

	// Check pool size
	if len(etp.standardTxs)+len(etp.enhancedTxs) >= etp.maxSize {
		return errors.New("transaction pool is full")
	}

	// Validate transaction
	if err := etp.validateStandardTransaction(tx); err != nil {
		return err
	}

	// Add transaction to pool
	etp.standardTxs[tx.Hash] = tx
	return nil
}

// AddEnhancedTransaction adds an enhanced transaction to the pool
func (etp *EnhancedTransactionPool) AddEnhancedTransaction(tx *EnhancedTransaction) error {
	etp.mu.Lock()
	defer etp.mu.Unlock()

	// Check pool size
	if len(etp.standardTxs)+len(etp.enhancedTxs) >= etp.maxSize {
		return errors.New("transaction pool is full")
	}

	// Validate enhanced transaction
	if err := etp.validateEnhancedTransaction(tx); err != nil {
		return err
	}

	// Add transaction to pool
	etp.enhancedTxs[tx.Hash] = tx
	return nil
}

// GetExecutableTransactions returns all transactions that can be executed
func (etp *EnhancedTransactionPool) GetExecutableTransactions() ([]*Transaction, []*EnhancedTransaction) {
	etp.mu.RLock()
	defer etp.mu.RUnlock()

	// Get all standard transactions
	standardTxs := make([]*Transaction, 0, len(etp.standardTxs))
	for _, tx := range etp.standardTxs {
		standardTxs = append(standardTxs, tx)
	}

	// Get executable enhanced transactions
	enhancedTxs := make([]*EnhancedTransaction, 0)
	for _, tx := range etp.enhancedTxs {
		if tx.IsExecutable() {
			enhancedTxs = append(enhancedTxs, tx)
		}
	}

	return standardTxs, enhancedTxs
}

// GetAllTransactions returns all transactions for backward compatibility
func (etp *EnhancedTransactionPool) GetAllTransactions() []*Transaction {
	etp.mu.RLock()
	defer etp.mu.RUnlock()

	allTxs := make([]*Transaction, 0, len(etp.standardTxs)+len(etp.enhancedTxs))

	// Add standard transactions
	for _, tx := range etp.standardTxs {
		allTxs = append(allTxs, tx)
	}

	// Add executable enhanced transactions converted to standard format
	for _, tx := range etp.enhancedTxs {
		if tx.IsExecutable() {
			standardTx := tx.ToStandardTransaction()
			allTxs = append(allTxs, &standardTx)
		}
	}

	return allTxs
}

// RemoveStandardTransactions removes standard transactions from the pool
func (etp *EnhancedTransactionPool) RemoveStandardTransactions(txs []*Transaction) {
	etp.mu.Lock()
	defer etp.mu.Unlock()

	for _, tx := range txs {
		delete(etp.standardTxs, tx.Hash)
	}
}

// RemoveEnhancedTransactions removes enhanced transactions from the pool
func (etp *EnhancedTransactionPool) RemoveEnhancedTransactions(txs []*EnhancedTransaction) {
	etp.mu.Lock()
	defer etp.mu.Unlock()

	for _, tx := range txs {
		delete(etp.enhancedTxs, tx.Hash)
	}
}

// GetPendingMultiSigTransactions returns multi-sig transactions pending signatures
func (etp *EnhancedTransactionPool) GetPendingMultiSigTransactions() []*EnhancedTransaction {
	etp.mu.RLock()
	defer etp.mu.RUnlock()

	pending := make([]*EnhancedTransaction, 0)
	for _, tx := range etp.enhancedTxs {
		if tx.Type == MultiSigTx && !tx.IsFullySigned() {
			pending = append(pending, tx)
		}
	}

	return pending
}

// GetTimeLockTransactions returns time-locked transactions (both ready and pending)
func (etp *EnhancedTransactionPool) GetTimeLockTransactions() (ready []*EnhancedTransaction, pending []*EnhancedTransaction) {
	etp.mu.RLock()
	defer etp.mu.RUnlock()

	for _, tx := range etp.enhancedTxs {
		if tx.Type == TimeLockTx {
			if tx.IsExecutable() {
				ready = append(ready, tx)
			} else {
				pending = append(pending, tx)
			}
		}
	}

	return ready, pending
}

// validateStandardTransaction validates a standard transaction
func (etp *EnhancedTransactionPool) validateStandardTransaction(tx *Transaction) error {
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
	if _, exists := etp.standardTxs[tx.Hash]; exists {
		return errors.New("transaction already exists in pool")
	}

	return nil
}

// validateEnhancedTransaction validates an enhanced transaction
func (etp *EnhancedTransactionPool) validateEnhancedTransaction(tx *EnhancedTransaction) error {
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
	if _, exists := etp.enhancedTxs[tx.Hash]; exists {
		return errors.New("transaction already exists in pool")
	}

	// Type-specific validation
	switch tx.Type {
	case MultiSigTx:
		if tx.RequiredSigs <= 0 || tx.RequiredSigs > len(tx.Signers) {
			return errors.New("invalid multi-sig transaction: invalid required signatures count")
		}
		if len(tx.Signers) == 0 {
			return errors.New("invalid multi-sig transaction: no signers specified")
		}
	case TimeLockTx:
		if tx.LockTime <= 0 {
			return errors.New("invalid time-lock transaction: invalid lock time")
		}
		if tx.LockTime <= time.Now().Unix() {
			return errors.New("invalid time-lock transaction: lock time must be in the future")
		}
	}

	return nil
}

// AddSignatureToTransaction adds a signature to a transaction in the pool
func (etp *EnhancedTransactionPool) AddSignatureToTransaction(txHash string, signature TransactionSignature) error {
	etp.mu.Lock()
	defer etp.mu.Unlock()

	tx, exists := etp.enhancedTxs[txHash]
	if !exists {
		return errors.New("transaction not found in pool")
	}

	return tx.AddSignature(signature)
}

// GetTransactionStats returns statistics about the transaction pool
func (etp *EnhancedTransactionPool) GetTransactionStats() map[string]int {
	etp.mu.RLock()
	defer etp.mu.RUnlock()

	stats := map[string]int{
		"standard_transactions": len(etp.standardTxs),
		"enhanced_transactions": len(etp.enhancedTxs),
		"total_transactions":    len(etp.standardTxs) + len(etp.enhancedTxs),
	}

	// Count enhanced transaction types
	multisig, timelock, contract, standard := 0, 0, 0, 0
	for _, tx := range etp.enhancedTxs {
		switch tx.Type {
		case MultiSigTx:
			multisig++
		case TimeLockTx:
			timelock++
		case ContractTx:
			contract++
		case StandardTx:
			standard++
		}
	}

	stats["multisig_transactions"] = multisig
	stats["timelock_transactions"] = timelock
	stats["contract_transactions"] = contract
	stats["enhanced_standard_transactions"] = standard

	return stats
}
