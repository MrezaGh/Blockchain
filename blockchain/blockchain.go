package blockchain

// Blockchain represents the blockchain
type Blockchain struct {
	Chain            []*Block
	Difficulty       int
	PendingTx        []Transaction
	MiningReward     float64
	MiningRewardAddr string
}

// NewBlockchain creates a new blockchain
func NewBlockchain(difficulty int, miningRewardAddr string) *Blockchain {
	bc := &Blockchain{
		Chain:            []*Block{createGenesisBlock()},
		Difficulty:       difficulty,
		PendingTx:        []Transaction{},
		MiningReward:     10.0,
		MiningRewardAddr: miningRewardAddr,
	}
	return bc
}

// createGenesisBlock creates the first block in the chain
func createGenesisBlock() *Block {
	return NewBlock(0, []Transaction{}, "0")
}

// GetLatestBlock returns the most recent block
func (bc *Blockchain) GetLatestBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}

// MinePendingTransactions mines pending transactions
func (bc *Blockchain) MinePendingTransactions() {
	// Create mining reward transaction
	rewardTx := Transaction{
		From:   "network",
		To:     bc.MiningRewardAddr,
		Amount: bc.MiningReward,
	}
	bc.PendingTx = append(bc.PendingTx, rewardTx)

	// Create new block
	block := NewBlock(
		int64(len(bc.Chain)),
		bc.PendingTx,
		bc.GetLatestBlock().Hash,
	)

	// Mine the block
	block.MineBlock(bc.Difficulty)

	// Add block to chain
	bc.Chain = append(bc.Chain, block)

	// Reset pending transactions
	bc.PendingTx = []Transaction{}
}

// AddTransaction adds a new transaction to pending transactions
func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.PendingTx = append(bc.PendingTx, tx)
}

// GetBalance calculates the balance of an address
func (bc *Blockchain) GetBalance(address string) float64 {
	var balance float64

	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			if tx.From == address {
				balance -= tx.Amount
			}
			if tx.To == address {
				balance += tx.Amount
			}
		}
	}

	return balance
}

// IsChainValid verifies if the blockchain is valid
func (bc *Blockchain) IsChainValid() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		// Verify current block's hash
		if currentBlock.Hash != currentBlock.calculateHash() {
			return false
		}

		// Verify chain linkage
		if currentBlock.PrevHash != previousBlock.Hash {
			return false
		}
	}

	return true
}
