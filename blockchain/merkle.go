package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root *MerkleNode
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  string
	Data  []byte
}

// NewMerkleTree creates a new Merkle tree from transaction data
func NewMerkleTree(transactions []Transaction) *MerkleTree {
	if len(transactions) == 0 {
		return &MerkleTree{Root: nil}
	}

	// Create leaf nodes from transactions
	var nodes []*MerkleNode
	for _, tx := range transactions {
		node := &MerkleNode{
			Hash: tx.Hash,
			Data: []byte(tx.Hash),
		}
		nodes = append(nodes, node)
	}

	// If odd number of transactions, duplicate the last one
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// Build the tree bottom-up
	for len(nodes) > 1 {
		var nextLevel []*MerkleNode

		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			right := nodes[i+1]

			parent := &MerkleNode{
				Left:  left,
				Right: right,
				Hash:  calculateNodeHash(left.Hash, right.Hash),
			}
			nextLevel = append(nextLevel, parent)
		}

		nodes = nextLevel
	}

	return &MerkleTree{Root: nodes[0]}
}

// calculateNodeHash calculates the hash of two child nodes
func calculateNodeHash(leftHash, rightHash string) string {
	data := leftHash + rightHash
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetMerkleRoot returns the root hash of the Merkle tree
func (mt *MerkleTree) GetMerkleRoot() string {
	if mt.Root == nil {
		return ""
	}
	return mt.Root.Hash
}

// MerkleProof represents a proof that a transaction exists in the tree
type MerkleProof struct {
	Hash   string   `json:"hash"`
	Hashes []string `json:"hashes"`
	IsLeft []bool   `json:"isLeft"` // Changed from Indices to IsLeft for clarity
}

// GenerateProof generates a Merkle proof for a given transaction hash
func (mt *MerkleTree) GenerateProof(txHash string) (*MerkleProof, error) {
	if mt.Root == nil {
		return nil, errors.New("empty tree")
	}

	proof := &MerkleProof{
		Hash:   txHash,
		Hashes: make([]string, 0),
		IsLeft: make([]bool, 0),
	}

	found := mt.buildProof(mt.Root, txHash, proof)
	if !found {
		return nil, errors.New("transaction not found in tree")
	}

	return proof, nil
}

// buildProof builds the proof path from leaf to root
func (mt *MerkleTree) buildProof(node *MerkleNode, txHash string, proof *MerkleProof) bool {
	if node == nil {
		return false
	}

	// If this is a leaf node, check if it matches our target
	if node.Left == nil && node.Right == nil {
		return node.Hash == txHash
	}

	// Check left subtree
	if node.Left != nil && mt.buildProof(node.Left, txHash, proof) {
		// Found in left subtree, add right sibling to proof
		if node.Right != nil {
			proof.Hashes = append(proof.Hashes, node.Right.Hash)
			proof.IsLeft = append(proof.IsLeft, false) // Right sibling, we are left
		}
		return true
	}

	// Check right subtree
	if node.Right != nil && mt.buildProof(node.Right, txHash, proof) {
		// Found in right subtree, add left sibling to proof
		if node.Left != nil {
			proof.Hashes = append(proof.Hashes, node.Left.Hash)
			proof.IsLeft = append(proof.IsLeft, true) // Left sibling, we are right
		}
		return true
	}

	return false
}

// VerifyProof verifies a Merkle proof against the root hash
func VerifyProof(proof *MerkleProof, rootHash string) bool {
	if len(proof.Hashes) != len(proof.IsLeft) {
		return false
	}

	currentHash := proof.Hash

	// Reconstruct the path to root (bottom-up)
	for i := 0; i < len(proof.Hashes); i++ {
		siblingHash := proof.Hashes[i]
		isLeft := proof.IsLeft[i]

		if isLeft {
			// Sibling is left, we are right
			currentHash = calculateNodeHash(siblingHash, currentHash)
		} else {
			// Sibling is right, we are left
			currentHash = calculateNodeHash(currentHash, siblingHash)
		}
	}

	return currentHash == rootHash
}

// GetTransactionHashes returns all transaction hashes in the tree (for debugging)
func (mt *MerkleTree) GetTransactionHashes() []string {
	if mt.Root == nil {
		return []string{}
	}

	var hashes []string
	mt.collectLeafHashes(mt.Root, &hashes)
	return hashes
}

// collectLeafHashes recursively collects all leaf node hashes
func (mt *MerkleTree) collectLeafHashes(node *MerkleNode, hashes *[]string) {
	if node == nil {
		return
	}

	// If this is a leaf node
	if node.Left == nil && node.Right == nil {
		*hashes = append(*hashes, node.Hash)
		return
	}

	// Recursively collect from children
	mt.collectLeafHashes(node.Left, hashes)
	mt.collectLeafHashes(node.Right, hashes)
}
