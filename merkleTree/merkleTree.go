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

type node struct {
	parent      *node
	childNumb   int
	children    []*node
	ownHash     [32]byte
	leaf        bool
	certificate []byte
	duplicate   bool
	id          int
}

type merkleTree struct {
	Root   *node
	leafs  []*node
	fanOut int
}

//Hello There Hello Hallo We are in

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}

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

	//var leafs []*node

	uneven := false
	if len(certs)%2 == 1 {
		certs = append(certs, certs[len(certs)-1])
		uneven = true
	}

	leafs := make([]*node, len(certs))

	for i := 0; i < len(certs); i++ {
		byteCet := certs[i]
		testHash := sha256.Sum256(byteCet)
		leafs[i] = &node{
			certificate: byteCet,
			childNumb:   i % 2,
			ownHash:     testHash,
			leaf:        true,
			duplicate:   false,
			id:          i,
		}
	}

	if uneven {
		leafs[len(leafs)-1].duplicate = true
		leafs[len(leafs)-1].childNumb = 1

	}

	nextLayer := makeLayer(leafs)
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer)
	}

	merk = merkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
	}

	return &merk
}

func makeLayer(nodes []*node) []*node {

	if len(nodes)%2 == 1 {
		appendNode := &node{
			certificate: nodes[len(nodes)-1].certificate,
			childNumb:   1 - nodes[len(nodes)-1].childNumb,
			ownHash:     nodes[len(nodes)-1].ownHash,
			children:    nodes[len(nodes)-1].children,
			leaf:        false,
			duplicate:   true,
			id:          nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)

	}

	nextLayer := make([]*node, len(nodes)/2) // divided with fanout which is 2 in this case

	for i := 0; i < len(nodes); {
		n1 := nodes[i]
		n2 := nodes[i+1]
		sum := sha256.Sum256(append(n1.ownHash[:], n2.ownHash[:]...))
		nextLayer[i/2] = &node{
			ownHash:   sum,
			childNumb: i % 2,
			leaf:      false,
			children:  []*node{n1, n2},
			id:        i / 2,
		}
		n1.parent = nextLayer[i/2]
		n2.parent = nextLayer[i/2]
		i = i + 2
	}
	return nextLayer
}

func verifyTree(certs [][]byte, tree merkleTree) bool {
	testTree := BuildTree(certs, tree.fanOut)

	return testTree.Root.ownHash == tree.Root.ownHash

}

func verifyNode(cert []byte, tree merkleTree) bool {
	var nod *node
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

	var hashList [][32]byte
	var childNumberList []int
	i := 0
	for nod.parent != nil {
		hashList = append(hashList, nod.parent.children[1-(nod.id%2)].ownHash)
		childNumberList = append(childNumberList, nod.id%2)
		nod = nod.parent
		i++
	}
	sum := sha256.Sum256(cert)
	for i := 0; i < len(hashList); i++ {
		if childNumberList[i]%2 == 0 {
			sum = sha256.Sum256(append(sum[:], hashList[i][:]...))
		} else {
			sum = sha256.Sum256(append(hashList[i][:], sum[:]...))
		}
	}
	return sum == tree.Root.ownHash

}
func updateTree() int {
	//TODO: Insert a node or delete a node?
	//HOw to do this, what is required??
	return 0
}

func main() {
	certArray := loadCertificates("testCerts/")
	updateTree()
	merkTree := BuildTree(certArray, 2)
	fmt.Println("Verify tree works for correct tree", verifyTree(certArray, *merkTree))
	fmt.Println("Verify node works for correct node", verifyNode(certArray[5], *merkTree))
	fmt.Println("Succes")

}
