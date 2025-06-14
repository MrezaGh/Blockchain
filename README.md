# Go Blockchain Implementation

A simple blockchain implementation in Go that demonstrates the core concepts of blockchain technology.

## Features

- Block creation and mining
- Proof of Work consensus
- Transaction management
- Wallet creation and management
- Cryptographic signatures
- Chain validation

## Project Structure

- `block.go`: Block structure and mining logic
- `blockchain.go`: Blockchain management and transaction handling
- `wallet.go`: Wallet creation and transaction signing
- `main.go`: Example usage of the blockchain

## Requirements

- Go 1.16 or higher

## Running the Example

```bash
go run main.go
```

## Implementation Details

### Block Structure
- Index
- Timestamp
- Transactions
- Previous hash
- Nonce
- Current hash

### Mining
- Proof of Work implementation
- Adjustable difficulty
- Mining rewards

### Transactions
- Transaction creation
- Transaction verification
- Balance tracking

### Security
- ECDSA signatures
- Chain validation
- Hash verification

## Example Usage

```go
// Create a new blockchain
bc := blockchain.NewBlockchain(4, "miner1")

// Create wallets
wallet1, _ := blockchain.NewWallet()
wallet2, _ := blockchain.NewWallet()

// Create and add transactions
tx := blockchain.Transaction{
    From:   wallet1.Address,
    To:     wallet2.Address,
    Amount: 10.0,
}
bc.AddTransaction(tx)

// Mine pending transactions
bc.MinePendingTransactions()

// Check balances
balance := bc.GetBalance(wallet1.Address)
```

## Note

This is a simplified implementation for educational purposes. It should not be used in production without additional security measures and optimizations. 