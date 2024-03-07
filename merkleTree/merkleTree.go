package main

import (
	"bytes"
	"crypto/sha256"

	//"errors"
	"fmt"
	//"hash"
	//"math/big"
	"log"
	"os"
)

// struct for representing a node in the tree
type node struct {
	parent      *node
	childNumb   int
	children    []*node
	ownHash     [32]byte
	certificate []byte
	duplicate   bool
	id          int
}

// Struct representing the merkle-tree
type merkleTree struct {
	Root   *node
	leafs  []*node
	fanOut int
}

// Function to call with error to avoid overloading methdods with error if statements
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}

// function to load certificates given, input which is the directory and amount represented as a list of ints,
// where [0} is the amount of certificates to load from said directory.
// returns a [][]byte list/array of files
func loadCertificates(input string, amount ...int) [][]byte {

	files, err := os.ReadDir(input)
	check(err)
	var fileArray [][]byte
	//To handle test cases, where we limit input size
	if len(amount) == 0 {
		fileArray = make([][]byte, len(files))
	} else {
		fileArray = make([][]byte, amount[0])
	}
	j := 0
	for i, v := range files {
		f, err := os.ReadFile("testCerts/" + v.Name())
		check(err)
		fileArray[i] = f
		j++
		if len(amount) > 0 && j == amount[0] {
			return fileArray
		}
	}
	return fileArray
}

func BuildTree(certs [][]byte, fanOut int) *merkleTree {
	var merk merkleTree

	leafs := make([]*node, len(certs))

	//build the leaf nodes of the tree
	for i := 0; i < len(certs); i++ {
		testHash := sha256.Sum256(certs[i])
		leafs[i] = &node{
			certificate: certs[i],
			childNumb:   i % fanOut,
			ownHash:     testHash,
			duplicate:   false,
			id:          i,
		}
	}

	//function call to make the next layer
	nextLayer := makeLayer(leafs, fanOut)
	// Checks if nextlayer it the root by checking the length, if not root call nextlayer again
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer, fanOut)
	}

	// define the merkletree struct
	merk = merkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
	}

	return &merk
}

func makeLayer(nodes []*node, fanOut int) []*node {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate: nodes[len(nodes)-1].certificate,
			childNumb:   (nodes[len(nodes)-1].id + 1) % fanOut,
			ownHash:     nodes[len(nodes)-1].ownHash,
			children:    nodes[len(nodes)-1].children,
			duplicate:   true,
			id:          nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)
	}

	nextLayer := make([]*node, len(nodes)/fanOut) // divided with fanout which is 2 in this case
	
	//The for loop which creates the next layer by create the vector commit for each of the new nodes.
	//And adding the corresponding children to each of their parents in the tree.
	for i := 0; i < len(nodes); {
		//The loop starts by finding the children for the current node in the 'nextlayer'
		var childrenList []*node
		var allChildrenHashes []byte
		for k := 0; k < fanOut; k++ {
			childrenList = append(childrenList, nodes[i+k])
			allChildrenHashes = append(allChildrenHashes, nodes[i+k].ownHash[:]...)
		}
		//Creates the hash of the children of the node.

		sum := sha256.Sum256(allChildrenHashes)
		//Creates the node with children and vectorcommit.
		nextLayer[i/fanOut] = &node{
			ownHash:   sum,
			childNumb: i % fanOut,
			children:  childrenList,
			id:        i / fanOut,
		}
		//Sets the parent for each of the nodes in the now previous layer.
		for _, v := range childrenList {
			v.parent = nextLayer[i/fanOut]
		}
		i = i + fanOut
	}
	return nextLayer
}

func verifyTree(certs [][]byte, tree merkleTree) bool {
	testTree := BuildTree(certs, tree.fanOut)

	return testTree.Root.ownHash == tree.Root.ownHash

}

func verifyNode(cert []byte, tree merkleTree) bool {
	var nod *node
	fanOut := tree.fanOut
	notInList := true
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, cert) {
			nod = v
			notInList = false
		}
	}
	if notInList {
		return false
	}

	var hashList [][][32]byte
	var childNumberList []int
	for nod.parent != nil {
		childNumberList = append(childNumberList, nod.id%fanOut)
		var hashList0 [][32]byte
		for _, v := range nod.parent.children {
			if nod.id != v.id {
				hashList0 = append(hashList0, v.ownHash)
			}
		}
		hashList = append(hashList, hashList0)
		nod = nod.parent
	}

	sum := sha256.Sum256(cert)
	for i := 0; i < len(hashList); i++ {
		var byteToHash []byte
		for j, v := range hashList[i] {

			if childNumberList[i] == j {
				byteToHash = append(byteToHash, sum[:]...)
			}
			byteToHash = append(byteToHash, v[:]...)
		}
		if childNumberList[i] == fanOut-1 {
			byteToHash = append(byteToHash, sum[:]...)
		}
		sum = sha256.Sum256(byteToHash)
	}
	return sum == tree.Root.ownHash
}

func updateLeaf(oldCert []byte, tree merkleTree, newCert []byte) *merkleTree {
	var nod *node
	notInList := true
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, oldCert) {
			nod = v
			notInList = false
		}
	}
	if notInList {
		return &tree
	}

	var childNumber int
	nod.certificate = newCert
	sum := sha256.Sum256(newCert)

	for nod.parent != nil {
		childNumber = nod.id % tree.fanOut
		var hashList [][32]byte
		for _, v := range nod.parent.children {
			if nod.id != v.id {
				hashList = append(hashList, v.ownHash)
			}
		}

		// Bla bla
		var byteToHash []byte
		for j, v := range hashList {
			if childNumber == j {
				byteToHash = append(byteToHash, sum[:]...)
			}
			byteToHash = append(byteToHash, v[:]...)
		}
		if childNumber == tree.fanOut-1 {
			byteToHash = append(byteToHash, sum[:]...)
		}
		sum = sha256.Sum256(byteToHash)
		nod.parent.ownHash = sum
		nod = nod.parent
	}
	return &tree
}

func insertLeaf(cert []byte, tree merkleTree) *merkleTree {
	//TODO: Insert a node or delete a node?
	//HOw to do this, what is required??
	return &tree
}

func deleteLeaf(cert []byte, tree merkleTree) *merkleTree {
	//TODO: Insert a node or delete a node?
	//How to do this, what is required??
	//Diego said this is not required and will be done when rebuilding or somewhere else by the CA
	fmt.Println("Sike! We cannot delete stuff.")
	return &tree
}

func loadOneCert(filePath string) []byte {
	f, err := os.ReadFile(filePath)
	check(err)
	return f
}

func main() {
	certArray := loadCertificates("testCerts/")
	merkTree := BuildTree(certArray, 2)
	fmt.Println("Verify tree works for correct tree", verifyTree(certArray, *merkTree))
	fmt.Println("Verify node works for correct node", verifyNode(certArray[5], *merkTree))
	updatedTree := updateLeaf(certArray[5], *merkTree, certArray[3])
	fmt.Println("We managed to overwrite a certificate", !verifyNode(certArray[5], *updatedTree))
	insertLeaf(certArray[5], *merkTree)
	deleteLeaf(certArray[5], *merkTree)

	fmt.Println("Succes")
}
