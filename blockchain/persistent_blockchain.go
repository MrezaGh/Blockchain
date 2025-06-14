package blockchain

import (
	"errors"
	"fmt"
	"log"
)

// PersistentBlockchain represents a blockchain with database persistence
type PersistentBlockchain struct {
	Chain            []*Block
	Difficulty       int
	TransactionPool  *TransactionPool
	EnhancedPool     *EnhancedTransactionPool
	MiningReward     float64
	MiningRewardAddr string
	Database         *Database
}

// NewPersistentBlockchain creates a new blockchain with database persistence
func NewPersistentBlockchain(difficulty int, miningRewardAddr string, dbConfig DatabaseConfig) (*PersistentBlockchain, error) {
	// Initialize database
	db, err := NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Try to load existing blockchain from database
	chain, err := db.LoadBlockchain()
	if err != nil {
		log.Printf("No existing blockchain found, creating new one: %v", err)
		// Create genesis block
		chain = []*Block{createGenesisBlock()}
	}

	// If no blocks loaded, create genesis block
	if len(chain) == 0 {
		chain = []*Block{createGenesisBlock()}
		// Save genesis block to database
		if err := db.SaveBlock(chain[0]); err != nil {
			log.Printf("Warning: failed to save genesis block: %v", err)
		}
	}

	pbc := &PersistentBlockchain{
		Chain:            chain,
		Difficulty:       difficulty,
		TransactionPool:  NewTransactionPool(1000),
		EnhancedPool:     NewEnhancedTransactionPool(1000),
		MiningReward:     10.0,
		MiningRewardAddr: miningRewardAddr,
		Database:         db,
	}

	log.Printf("Loaded blockchain with %d blocks from database", len(chain))
	return pbc, nil
}

// Close closes the blockchain and database connections
func (pbc *PersistentBlockchain) Close() error {
	return pbc.Database.Close()
}

// GetLatestBlock returns the most recent block
func (pbc *PersistentBlockchain) GetLatestBlock() *Block {
	return pbc.Chain[len(pbc.Chain)-1]
}

// MinePendingTransactions mines pending transactions and persists the new block
func (pbc *PersistentBlockchain) MinePendingTransactions() error {
	// Create mining reward transaction
	rewardTx := NewTransaction("network", pbc.MiningRewardAddr, pbc.MiningReward, 0)
	pbc.TransactionPool.AddTransaction(rewardTx)

	// Get transactions from pool
	pendingTxs := pbc.TransactionPool.GetTransactions()

	// Also get executable enhanced transactions
	_, enhancedTxs := pbc.EnhancedPool.GetExecutableTransactions()

	// Convert enhanced transactions to standard format for block inclusion
	for _, eTx := range enhancedTxs {
		standardTx := eTx.ToStandardTransaction()
		pendingTxs = append(pendingTxs, &standardTx)
	}

	// Convert []*Transaction to []Transaction
	transactions := make([]Transaction, len(pendingTxs))
	for i, tx := range pendingTxs {
		transactions[i] = *tx
	}

	// Create new block
	block := NewBlock(
		int64(len(pbc.Chain)),
		transactions,
		pbc.GetLatestBlock().Hash,
	)

	// Mine the block
	log.Printf("Mining block %d with %d transactions...", block.Index, len(transactions))
	block.MineBlock(pbc.Difficulty)

	// Add block to chain
	pbc.Chain = append(pbc.Chain, block)

	// Save block to database
	if err := pbc.Database.SaveBlock(block); err != nil {
		log.Printf("Error saving block to database: %v", err)
		// Remove block from chain if database save failed
		pbc.Chain = pbc.Chain[:len(pbc.Chain)-1]
		return fmt.Errorf("failed to persist block: %v", err)
	}

	// Remove mined transactions from pools
	pbc.TransactionPool.RemoveTransactions(pendingTxs)
	pbc.EnhancedPool.RemoveEnhancedTransactions(enhancedTxs)

	log.Printf("Block %d mined and persisted successfully", block.Index)
	return nil
}

// AddTransaction adds a new transaction to the transaction pool
func (pbc *PersistentBlockchain) AddTransaction(tx *Transaction) error {
	return pbc.TransactionPool.AddTransaction(tx)
}

// AddEnhancedTransaction adds a new enhanced transaction to the enhanced pool
func (pbc *PersistentBlockchain) AddEnhancedTransaction(tx *EnhancedTransaction) error {
	return pbc.EnhancedPool.AddEnhancedTransaction(tx)
}

// GetBalance calculates the balance of an address (from database for better performance)
func (pbc *PersistentBlockchain) GetBalance(address string) float64 {
	// Try to get balance from database first (more efficient)
	balance, err := pbc.Database.GetAddressBalance(address)
	if err != nil {
		log.Printf("Error getting balance from database, calculating from chain: %v", err)
		// Fallback to chain calculation
		return pbc.calculateBalanceFromChain(address)
	}
	return balance
}

// calculateBalanceFromChain calculates balance by iterating through the chain (fallback method)
func (pbc *PersistentBlockchain) calculateBalanceFromChain(address string) float64 {
	var balance float64

	for _, block := range pbc.Chain {
		for _, tx := range block.Transactions {
			if tx.From == address {
				balance -= tx.Amount + tx.Fee
			}
			if tx.To == address {
				balance += tx.Amount
			}
		}
	}

	return balance
}

// IsChainValid verifies if the blockchain is valid
func (pbc *PersistentBlockchain) IsChainValid() bool {
	for i := 1; i < len(pbc.Chain); i++ {
		currentBlock := pbc.Chain[i]
		previousBlock := pbc.Chain[i-1]

		// Verify current block's hash
		if currentBlock.Hash != currentBlock.calculateHash() {
			log.Printf("Invalid hash at block %d", i)
			return false
		}

		// Verify chain linkage
		if currentBlock.PrevHash != previousBlock.Hash {
			log.Printf("Invalid chain linkage at block %d", i)
			return false
		}

		// Verify Merkle tree integrity
		if !currentBlock.ValidateTransactions() {
			log.Printf("Invalid Merkle tree at block %d", i)
			return false
		}
	}

	return true
}

// GetTransactionProof generates a Merkle proof for a transaction in a specific block
func (pbc *PersistentBlockchain) GetTransactionProof(blockIndex int, txHash string) (*MerkleProof, error) {
	if blockIndex < 0 || blockIndex >= len(pbc.Chain) {
		return nil, errors.New("invalid block index")
	}

	block := pbc.Chain[blockIndex]
	return block.GenerateTransactionProof(txHash)
}

// VerifyTransactionInBlock verifies that a transaction exists in a specific block
func (pbc *PersistentBlockchain) VerifyTransactionInBlock(blockIndex int, proof *MerkleProof) bool {
	if blockIndex < 0 || blockIndex >= len(pbc.Chain) {
		return false
	}

	block := pbc.Chain[blockIndex]
	return block.VerifyTransactionProof(proof)
}

// GetBlockchainStats returns comprehensive blockchain statistics
func (pbc *PersistentBlockchain) GetBlockchainStats() (map[string]interface{}, error) {
	// Get stats from database
	dbStats, err := pbc.Database.GetBlockchainStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %v", err)
	}

	// Add memory pool stats
	dbStats["pending_transactions"] = len(pbc.TransactionPool.GetTransactions())
	dbStats["pending_enhanced_transactions"] = len(pbc.EnhancedPool.GetAllTransactions())

	// Add enhanced transaction pool stats
	enhancedStats := pbc.EnhancedPool.GetTransactionStats()
	for key, value := range enhancedStats {
		dbStats["pool_"+key] = value
	}

	// Add chain validation status
	dbStats["chain_valid"] = pbc.IsChainValid()
	dbStats["in_memory_blocks"] = len(pbc.Chain)

	return dbStats, nil
}

// RecoverFromDatabase recovers the blockchain state from database
func (pbc *PersistentBlockchain) RecoverFromDatabase() error {
	log.Println("Recovering blockchain from database...")

	// Load blockchain from database
	chain, err := pbc.Database.LoadBlockchain()
	if err != nil {
		return fmt.Errorf("failed to load blockchain from database: %v", err)
	}

	if len(chain) == 0 {
		return errors.New("no blocks found in database")
	}

	// Validate the loaded chain
	tempBC := &PersistentBlockchain{Chain: chain}
	if !tempBC.IsChainValid() {
		return errors.New("loaded blockchain is invalid")
	}

	// Update the current blockchain
	pbc.Chain = chain

	log.Printf("Successfully recovered blockchain with %d blocks", len(chain))
	return nil
}

// SyncWithDatabase ensures the in-memory chain matches the database
func (pbc *PersistentBlockchain) SyncWithDatabase() error {
	log.Println("Syncing blockchain with database...")

	// Get latest block from database
	latestDBBlock, err := pbc.Database.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("failed to get latest block from database: %v", err)
	}

	// Compare with in-memory chain
	latestMemoryBlock := pbc.GetLatestBlock()

	if latestDBBlock.Index != latestMemoryBlock.Index {
		log.Printf("Blockchain out of sync. DB: %d, Memory: %d", latestDBBlock.Index, latestMemoryBlock.Index)
		return pbc.RecoverFromDatabase()
	}

	if latestDBBlock.Hash != latestMemoryBlock.Hash {
		log.Printf("Hash mismatch at block %d", latestDBBlock.Index)
		return pbc.RecoverFromDatabase()
	}

	log.Println("Blockchain is in sync with database")
	return nil
}

// BackupBlockchain creates a backup of the current blockchain state
func (pbc *PersistentBlockchain) BackupBlockchain(backupPath string) error {
	// This would implement blockchain backup functionality
	// For now, it's a placeholder
	log.Printf("Backup functionality would save blockchain to: %s", backupPath)
	return nil
}

// GetBlockByHash retrieves a block by its hash (from database)
func (pbc *PersistentBlockchain) GetBlockByHash(hash string) (*Block, error) {
	return pbc.Database.GetBlock(hash)
}

// GetBlockByIndex retrieves a block by its index (from database)
func (pbc *PersistentBlockchain) GetBlockByIndex(index int64) (*Block, error) {
	return pbc.Database.GetBlockByIndex(index)
}
