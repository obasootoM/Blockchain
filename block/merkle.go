package block

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHash)
		node.Data = hash[:]
	}
	node.Right = right
	node.Left = left
	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var node []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, datas := range data {
		nodes := NewMerkleNode(nil, nil, datas)
		node = append(node, *nodes)
	}
	for i := 0; i < len(data)/2; i++ {
		var level []MerkleNode
		for j := 0; j < len(node); j += 2 {
			nodes := NewMerkleNode(&node[j], &node[j+1], nil)
			level = append(level, *nodes)
		}
		node = level
	}
	tree := MerkleTree{&node[0]}
	return &tree
}
