package main

import (
	"fmt"
	"log"

	"blockchain/blockchain"
)

func main() {
	// Create a new blockchain with difficulty 4
	bc := blockchain.NewBlockchain(4, "miner1")

	// Create two wallets
	wallet1, err := blockchain.NewWallet()
	if err != nil {
		log.Fatal(err)
	}

	wallet2, err := blockchain.NewWallet()
	if err != nil {
		log.Fatal(err)
	}

	// Create some transactions
	tx1 := blockchain.Transaction{
		From:   wallet1.Address,
		To:     wallet2.Address,
		Amount: 10.0,
	}

	tx2 := blockchain.Transaction{
		From:   wallet2.Address,
		To:     wallet1.Address,
		Amount: 5.0,
	}

	// Add transactions to the blockchain
	bc.AddTransaction(tx1)
	bc.AddTransaction(tx2)

	// Mine pending transactions
	fmt.Println("Mining block 1...")
	bc.MinePendingTransactions()

	// Print balances
	fmt.Printf("Wallet 1 balance: %.2f\n", bc.GetBalance(wallet1.Address))
	fmt.Printf("Wallet 2 balance: %.2f\n", bc.GetBalance(wallet2.Address))

	// Verify the chain
	fmt.Printf("Is chain valid? %v\n", bc.IsChainValid())

	// Print blockchain info
	fmt.Printf("Number of blocks: %d\n", len(bc.Chain))
	fmt.Printf("Latest block hash: %s\n", bc.GetLatestBlock().Hash)
}
