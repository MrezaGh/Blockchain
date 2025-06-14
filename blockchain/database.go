package blockchain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the blockchain database
type Database struct {
	db   *sql.DB
	path string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string
	Path     string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*Database, error) {
	var db *sql.DB
	var err error

	switch config.Driver {
	case "sqlite3":
		db, err = sql.Open("sqlite3", config.Path)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.DBName)
		db, err = sql.Open("postgres", dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	database := &Database{
		db:   db,
		path: config.Path,
	}

	// Initialize database schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initSchema initializes the database schema
func (d *Database) initSchema() error {
	// Create blocks table
	blocksTable := `
	CREATE TABLE IF NOT EXISTS blocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		block_index INTEGER UNIQUE NOT NULL,
		hash TEXT UNIQUE NOT NULL,
		previous_hash TEXT NOT NULL,
		merkle_root TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		nonce INTEGER NOT NULL,
		difficulty INTEGER NOT NULL,
		transaction_count INTEGER NOT NULL,
		block_data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create transactions table
	transactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT UNIQUE NOT NULL,
		block_hash TEXT NOT NULL,
		block_index INTEGER NOT NULL,
		tx_index INTEGER NOT NULL,
		from_address TEXT NOT NULL,
		to_address TEXT NOT NULL,
		amount REAL NOT NULL,
		fee REAL NOT NULL,
		timestamp INTEGER NOT NULL,
		transaction_data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(block_hash) REFERENCES blocks(hash)
	);`

	// Create enhanced transactions table
	enhancedTransactionsTable := `
	CREATE TABLE IF NOT EXISTS enhanced_transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		transaction_id TEXT UNIQUE NOT NULL,
		hash TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		from_address TEXT NOT NULL,
		to_address TEXT NOT NULL,
		amount REAL NOT NULL,
		fee REAL NOT NULL,
		timestamp INTEGER NOT NULL,
		required_sigs INTEGER DEFAULT 0,
		current_sigs INTEGER DEFAULT 0,
		lock_time INTEGER DEFAULT 0,
		is_executed BOOLEAN DEFAULT FALSE,
		transaction_data TEXT NOT NULL,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create addresses table for balance indexing
	addressesTable := `
	CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT UNIQUE NOT NULL,
		balance REAL DEFAULT 0.0,
		transaction_count INTEGER DEFAULT 0,
		first_seen INTEGER NOT NULL,
		last_updated INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create blockchain state table
	blockchainStateTable := `
	CREATE TABLE IF NOT EXISTS blockchain_state (
		id INTEGER PRIMARY KEY,
		latest_block_hash TEXT NOT NULL,
		latest_block_index INTEGER NOT NULL,
		total_blocks INTEGER NOT NULL,
		total_transactions INTEGER NOT NULL,
		difficulty INTEGER NOT NULL,
		mining_reward REAL NOT NULL,
		last_updated INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_blocks_index ON blocks(block_index);",
		"CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash);",
		"CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);",
		"CREATE INDEX IF NOT EXISTS idx_transactions_block ON transactions(block_hash);",
		"CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_address);",
		"CREATE INDEX IF NOT EXISTS idx_transactions_to ON transactions(to_address);",
		"CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_enhanced_transactions_type ON enhanced_transactions(type);",
		"CREATE INDEX IF NOT EXISTS idx_enhanced_transactions_from ON enhanced_transactions(from_address);",
		"CREATE INDEX IF NOT EXISTS idx_enhanced_transactions_to ON enhanced_transactions(to_address);",
		"CREATE INDEX IF NOT EXISTS idx_addresses_address ON addresses(address);",
	}

	// Execute table creation statements
	tables := []string{blocksTable, transactionsTable, enhancedTransactionsTable, addressesTable, blockchainStateTable}

	for _, table := range tables {
		if _, err := d.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	// Create indexes
	for _, index := range indexes {
		if _, err := d.db.Exec(index); err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	return nil
}

// SaveBlock saves a block to the database
func (d *Database) SaveBlock(block *Block) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Serialize block data
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to serialize block: %v", err)
	}

	// Insert block
	_, err = tx.Exec(`
		INSERT INTO blocks (block_index, hash, previous_hash, merkle_root, timestamp, nonce, difficulty, transaction_count, block_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		block.Index, block.Hash, block.PrevHash, block.MerkleRoot,
		block.Timestamp, block.Nonce, 4, // difficulty hardcoded for now
		len(block.Transactions), string(blockData))

	if err != nil {
		return fmt.Errorf("failed to insert block: %v", err)
	}

	// Save transactions
	for i, transaction := range block.Transactions {
		if err := d.saveTransaction(tx, &transaction, block.Hash, block.Index, i); err != nil {
			return fmt.Errorf("failed to save transaction: %v", err)
		}
	}

	// Update blockchain state
	if err := d.updateBlockchainState(tx, block); err != nil {
		return fmt.Errorf("failed to update blockchain state: %v", err)
	}

	return tx.Commit()
}

// saveTransaction saves a transaction to the database (internal helper)
func (d *Database) saveTransaction(tx *sql.Tx, transaction *Transaction, blockHash string, blockIndex int64, txIndex int) error {
	// Serialize transaction data
	txData, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %v", err)
	}

	// Insert transaction
	_, err = tx.Exec(`
		INSERT INTO transactions (hash, block_hash, block_index, tx_index, from_address, to_address, amount, fee, timestamp, transaction_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		transaction.Hash, blockHash, blockIndex, txIndex,
		transaction.From, transaction.To, transaction.Amount, transaction.Fee,
		time.Now().Unix(), string(txData))

	if err != nil {
		return err
	}

	// Update address balances
	if err := d.updateAddressBalance(tx, transaction.From, -transaction.Amount-transaction.Fee); err != nil {
		return err
	}
	if err := d.updateAddressBalance(tx, transaction.To, transaction.Amount); err != nil {
		return err
	}

	return nil
}

// updateAddressBalance updates the balance for an address
func (d *Database) updateAddressBalance(tx *sql.Tx, address string, change float64) error {
	now := time.Now().Unix()

	// Try to update existing address
	result, err := tx.Exec(`
		UPDATE addresses SET balance = balance + ?, transaction_count = transaction_count + 1, last_updated = ?
		WHERE address = ?`, change, now, address)
	if err != nil {
		return err
	}

	// If no rows affected, insert new address
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		_, err = tx.Exec(`
			INSERT INTO addresses (address, balance, transaction_count, first_seen, last_updated)
			VALUES (?, ?, 1, ?, ?)`, address, change, now, now)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateBlockchainState updates the blockchain state
func (d *Database) updateBlockchainState(tx *sql.Tx, block *Block) error {
	now := time.Now().Unix()

	// Try to update existing state
	result, err := tx.Exec(`
		UPDATE blockchain_state SET 
			latest_block_hash = ?, 
			latest_block_index = ?, 
			total_blocks = total_blocks + 1, 
			total_transactions = total_transactions + ?, 
			last_updated = ?
		WHERE id = 1`, block.Hash, block.Index, len(block.Transactions), now)

	if err != nil {
		return err
	}

	// If no rows affected, insert initial state
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		_, err = tx.Exec(`
			INSERT INTO blockchain_state (id, latest_block_hash, latest_block_index, total_blocks, total_transactions, difficulty, mining_reward, last_updated)
			VALUES (1, ?, ?, 1, ?, 4, 10.0, ?)`,
			block.Hash, block.Index, len(block.Transactions), now)
	}

	return err
}

// GetBlock retrieves a block by hash
func (d *Database) GetBlock(hash string) (*Block, error) {
	var blockData string
	err := d.db.QueryRow("SELECT block_data FROM blocks WHERE hash = ?", hash).Scan(&blockData)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal([]byte(blockData), &block); err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %v", err)
	}

	return &block, nil
}

// GetBlockByIndex retrieves a block by index
func (d *Database) GetBlockByIndex(index int64) (*Block, error) {
	var blockData string
	err := d.db.QueryRow("SELECT block_data FROM blocks WHERE block_index = ?", index).Scan(&blockData)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal([]byte(blockData), &block); err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %v", err)
	}

	return &block, nil
}

// GetLatestBlock retrieves the latest block
func (d *Database) GetLatestBlock() (*Block, error) {
	var blockData string
	err := d.db.QueryRow("SELECT block_data FROM blocks ORDER BY block_index DESC LIMIT 1").Scan(&blockData)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal([]byte(blockData), &block); err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %v", err)
	}

	return &block, nil
}

// GetAddressBalance retrieves the balance for an address
func (d *Database) GetAddressBalance(address string) (float64, error) {
	var balance float64
	err := d.db.QueryRow("SELECT COALESCE(balance, 0) FROM addresses WHERE address = ?", address).Scan(&balance)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return balance, nil
}

// GetBlockchainStats retrieves blockchain statistics
func (d *Database) GetBlockchainStats() (map[string]interface{}, error) {
	var stats = make(map[string]interface{})

	// Get basic stats from blockchain_state table
	var latestBlockHash string
	var latestBlockIndex, totalBlocks, totalTransactions int64
	var difficulty int
	var miningReward float64
	var lastUpdated int64

	err := d.db.QueryRow(`
		SELECT latest_block_hash, latest_block_index, total_blocks, total_transactions, 
		       difficulty, mining_reward, last_updated 
		FROM blockchain_state WHERE id = 1`).Scan(
		&latestBlockHash, &latestBlockIndex, &totalBlocks, &totalTransactions,
		&difficulty, &miningReward, &lastUpdated)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	stats["latest_block_hash"] = latestBlockHash
	stats["latest_block_index"] = latestBlockIndex
	stats["total_blocks"] = totalBlocks
	stats["total_transactions"] = totalTransactions
	stats["difficulty"] = difficulty
	stats["mining_reward"] = miningReward
	stats["last_updated"] = lastUpdated

	// Get additional stats
	var addressCount int64
	d.db.QueryRow("SELECT COUNT(*) FROM addresses").Scan(&addressCount)
	stats["total_addresses"] = addressCount

	return stats, nil
}

// LoadBlockchain loads the entire blockchain from database
func (d *Database) LoadBlockchain() ([]*Block, error) {
	rows, err := d.db.Query("SELECT block_data FROM blocks ORDER BY block_index ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*Block
	for rows.Next() {
		var blockData string
		if err := rows.Scan(&blockData); err != nil {
			return nil, err
		}

		var block Block
		if err := json.Unmarshal([]byte(blockData), &block); err != nil {
			return nil, fmt.Errorf("failed to deserialize block: %v", err)
		}

		blocks = append(blocks, &block)
	}

	return blocks, nil
}
