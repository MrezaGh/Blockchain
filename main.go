package main

import (
	"fmt"
	"log"

	"blockchain/blockchain"
)

func main() {
	fmt.Println("=== Enhanced Blockchain with Merkle Trees ===\n")

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
	tx1 := blockchain.NewTransaction(wallet1.Address, wallet2.Address, 10.0, 0.1)
	tx2 := blockchain.NewTransaction(wallet2.Address, wallet1.Address, 5.0, 0.1)
	tx3 := blockchain.NewTransaction(wallet1.Address, wallet2.Address, 3.0, 0.1)

	// Add transactions to the blockchain
	if err := bc.AddTransaction(tx1); err != nil {
		log.Printf("Error adding transaction 1: %v", err)
	}
	if err := bc.AddTransaction(tx2); err != nil {
		log.Printf("Error adding transaction 2: %v", err)
	}
	if err := bc.AddTransaction(tx3); err != nil {
		log.Printf("Error adding transaction 3: %v", err)
	}

	// Mine pending transactions
	fmt.Println("Mining block 1...")
	bc.MinePendingTransactions()

	// Print balances
	fmt.Printf("Wallet 1 balance: %.2f\n", bc.GetBalance(wallet1.Address))
	fmt.Printf("Wallet 2 balance: %.2f\n", bc.GetBalance(wallet2.Address))

	// Verify the chain (now includes Merkle tree validation)
	fmt.Printf("Is chain valid? %v\n", bc.IsChainValid())

	// Print blockchain info
	fmt.Printf("Number of blocks: %d\n", len(bc.Chain))
	fmt.Printf("Latest block hash: %s\n", bc.GetLatestBlock().Hash)
	fmt.Printf("Latest block Merkle root: %s\n", bc.GetLatestBlock().MerkleRoot)

	// Demonstrate Merkle proof functionality
	fmt.Println("\n=== Merkle Proof Demonstration ===")

	latestBlock := bc.GetLatestBlock()
	if len(latestBlock.Transactions) > 0 {
		// Generate proof for the first transaction
		txHash := latestBlock.Transactions[0].Hash
		fmt.Printf("Generating proof for transaction: %s\n", txHash[:16]+"...")

		proof, err := bc.GetTransactionProof(len(bc.Chain)-1, txHash)
		if err != nil {
			log.Printf("Error generating proof: %v", err)
		} else {
			fmt.Printf("Proof generated successfully with %d hashes\n", len(proof.Hashes))

			// Verify the proof
			isValid := bc.VerifyTransactionInBlock(len(bc.Chain)-1, proof)
			fmt.Printf("Proof verification result: %v\n", isValid)

			// Demonstrate light client verification (without full block data)
			isValidDirect := blockchain.VerifyProof(proof, latestBlock.MerkleRoot)
			fmt.Printf("Direct proof verification: %v\n", isValidDirect)
		}
	}

	// Add more transactions and mine another block
	fmt.Println("\n=== Mining Second Block ===")

	tx4 := blockchain.NewTransaction(wallet1.Address, wallet2.Address, 7.0, 0.1)
	tx5 := blockchain.NewTransaction(wallet2.Address, wallet1.Address, 2.0, 0.1)

	bc.AddTransaction(tx4)
	bc.AddTransaction(tx5)

	fmt.Println("Mining block 2...")
	bc.MinePendingTransactions()

	// Final verification
	fmt.Printf("Final chain validation: %v\n", bc.IsChainValid())
	fmt.Printf("Total blocks: %d\n", len(bc.Chain))

	// Print final balances
	fmt.Printf("Final Wallet 1 balance: %.2f\n", bc.GetBalance(wallet1.Address))
	fmt.Printf("Final Wallet 2 balance: %.2f\n", bc.GetBalance(wallet2.Address))

	fmt.Println("\n=== Enhancement 1 Complete: Merkle Trees Implemented ===")
}
