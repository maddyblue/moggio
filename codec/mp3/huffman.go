// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mp3

import (
	"fmt"
	"sort"
)

// A huffmanTree is a binary tree which is navigated, bit-by-bit to reach a
// symbol.
type huffmanTree struct {
	// nodes contains all the non-leaf nodes in the tree. nodes[0] is the
	// root of the tree and nextNode contains the index of the next element
	// of nodes to use when the tree is being constructed.
	nodes    []huffmanNode
	nextNode int
}

// A huffmanNode is a node in the tree. left and right contain indexes into the
// nodes slice of the tree. If left or right is invalidNodeValue then the child
// is a left node and its value is in leftValue/rightValue.
type huffmanNode struct {
	left, right           uint16
	leftValue, rightValue [2]byte
}

// invalidNodeValue is an invalid index which marks a leaf node in the tree.
const invalidNodeValue = 0xffff

// Decode reads bits from the given bitReader and navigates the tree until a
// symbol is found.
func (t *huffmanTree) Decode(br *bitReader) (v [2]byte) {
	nodeIndex := uint16(0) // node 0 is the root of the tree.

	for {
		node := &t.nodes[nodeIndex]
		bit, ok := br.TryReadBit()
		if !ok && br.ReadBit() {
			bit = 1
		}
		// mp3 encodes right as a true bit.
		if bit == 0 {
			// left
			if node.left == invalidNodeValue {
				return node.leftValue
			}
			nodeIndex = node.left
		} else {
			// right
			if node.right == invalidNodeValue {
				return node.rightValue
			}
			nodeIndex = node.right
		}
	}
}

func mustHuffmanTree(pairs []huffmanPair) huffmanTree {
	t, err := newHuffmanTree(pairs)
	if err != nil {
		panic(err)
	}
	return t
}

func newHuffmanTree(pairs []huffmanPair) (huffmanTree, error) {
	if len(pairs) < 2 {
		panic("newHuffmanTree: too few symbols")
	}

	var t huffmanTree

	// Now we assign codes to the symbols, starting with the longest code.
	// We keep the codes packed into a uint32, at the most-significant end.
	// So branches are taken from the MSB downwards. This makes it easy to
	// sort them later.

	codes := huffmanCodes(make([]huffmanCode, len(pairs)))
	for i, pair := range pairs {
		codes[i].codeLen = uint8(len(pair.bits))
		codes[i].value = pair.value
		var code uint32
		for _, b := range pair.bits {
			code <<= 1
			code += uint32(b)
		}
		code <<= 32 - codes[i].codeLen
		codes[i].code = code
	}

	// Now we can sort by the code so that the left half of each branch are
	// grouped together, recursively.
	sort.Sort(codes)

	t.nodes = make([]huffmanNode, len(codes))
	_, err := buildHuffmanNode(&t, codes, 0)
	return t, err
}

type huffmanPair struct {
	bits  []byte
	value [2]byte
}

// huffmanCode contains a symbol, its code and code length.
type huffmanCode struct {
	code    uint32
	codeLen uint8
	value   [2]byte
}

// huffmanCodes is used to provide an interface for sorting.
type huffmanCodes []huffmanCode

func (n huffmanCodes) Len() int {
	return len(n)
}

func (n huffmanCodes) Less(i, j int) bool {
	return n[i].code < n[j].code
}

func (n huffmanCodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// buildHuffmanNode takes a slice of sorted huffmanCodes and builds a node in
// the Huffman tree at the given level. It returns the index of the newly
// constructed node.
func buildHuffmanNode(t *huffmanTree, codes []huffmanCode, level uint32) (nodeIndex uint16, err error) {
	test := uint32(1) << (31 - level)

	// We have to search the list of codes to find the divide between the left and right sides.
	firstRightIndex := len(codes)
	for i, code := range codes {
		if code.code&test != 0 {
			firstRightIndex = i
			break
		}
	}

	left := codes[:firstRightIndex]
	right := codes[firstRightIndex:]

	if len(left) == 0 || len(right) == 0 {
		return 0, fmt.Errorf("superfluous level in Huffman tree")
	}

	nodeIndex = uint16(t.nextNode)
	node := &t.nodes[t.nextNode]
	t.nextNode++

	if len(left) == 1 {
		// leaf node
		node.left = invalidNodeValue
		node.leftValue = left[0].value
	} else {
		node.left, err = buildHuffmanNode(t, left, level+1)
	}

	if err != nil {
		return
	}

	if len(right) == 1 {
		// leaf node
		node.right = invalidNodeValue
		node.rightValue = right[0].value
	} else {
		node.right, err = buildHuffmanNode(t, right, level+1)
	}
	return
}
