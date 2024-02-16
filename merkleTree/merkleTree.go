package main

import (
	"bytes"
	"crypto/sha256"

	//"errors"
	"fmt"
	//"hash"
	//"math/big"
	"reflect"
)

type node struct {
	parent      *node
	childNumb   int
	children    []*node
	ownHash     [32]byte
	leaf        bool
	certificate []byte
	duplicate   bool
}

type merkleTree struct {
	Root   *node
	leafs  []*node
	fanOut int
}

func BuildTree(certs []string, fanOut int) *merkleTree {
	var merk merkleTree

	//var leafs []*node

	uneven := false
	if len(certs)%2 == 1 {
		certs = append(certs, certs[len(certs)])
		uneven = true
	}

	leafs := make([]*node, len(certs))

	for i := 0; i < len(certs); i++ {
		byteCet := []byte(certs[i])
		testHash := sha256.Sum256(byteCet)
		leafs[i] = &node{
			certificate: byteCet,
			childNumb:   i % 2,
			ownHash:     testHash,
			leaf:        true,
			duplicate:   false,
		}
	}

	if uneven {
		leafs[len(leafs)].duplicate = true
	}

	fmt.Println(reflect.TypeOf(leafs))

	nextLayer := makeLayer(leafs)
	fmt.Println(len(nextLayer))
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer)
		fmt.Println(len(nextLayer))
	}

	merk = merkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
	}

	return &merk
}

func makeLayer(nodes []*node) []*node {

	nextLayer := make([]*node, len(nodes)/2) // divided with fanout which is 2 in this case

	for i := 0; i < len(nodes); {
		n1 := nodes[i]
		n2 := nodes[i+1]
		fmt.Println(n2.ownHash[:])
		sum := sha256.Sum256(append(n1.ownHash[:], n2.ownHash[:]...))
		nextLayer[i/2] = &node{
			ownHash:   sum,
			childNumb: i % 2,
			leaf:      false,
			children:  []*node{n1, n2},
		}
		n1.parent = nextLayer[i/2]
		n2.parent = nextLayer[i/2]
		i = i + 2
	}

	return nextLayer
}

func verifyTree(certs []string, tree merkleTree) bool {
	testTree := BuildTree(certs, tree.fanOut)
	return testTree.Root.ownHash == tree.Root.ownHash

}

func verifyNode(cert string, tree merkleTree) bool {
	var nod *node
	notInList := true
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, []byte(cert)) {
			nod = v
			notInList = false
		}
	}
	if notInList {
		return false
	}

	var hashList [][32]byte
	var childNumberList []int
	i := 0
	fmt.Println(nod)
	for nod.parent != nil {
		fmt.Println("We made it here1")
		hashList = append(hashList, nod.parent.children[1-nod.childNumb].ownHash)
		childNumberList = append(childNumberList, nod.childNumb)
		nod = nod.parent
		i++
	}
	fmt.Println("We made it here2")
	sum := sha256.Sum256([]byte(cert))
	for i := 0; i < len(hashList); i++ {
		if childNumberList[i]%2 == 0 {
			sum = sha256.Sum256(append(sum[:], hashList[i][:]...))
		} else {
			sum = sha256.Sum256(append(hashList[i][:], sum[:]...))
		}

	}
	return sum == tree.Root.ownHash

}

func main() {
	l := []string{"gedshsfhdfghfghd", "jfdghfghfdghfghens", "dortfghfdghfdghfdgjhe", "fledfhfgjfdjdfgjdfghmming"}
	merkTree := BuildTree(l, 2)
	fmt.Println(merkTree)
	fmt.Println(verifyTree(l, *merkTree))
	fmt.Println("Succes")
	fmt.Println("Certificate is in tree: ", verifyNode("gedshsfhsdfghfghd", *merkTree))
}
