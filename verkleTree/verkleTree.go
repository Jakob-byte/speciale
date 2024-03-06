package verkletree

import (
	"bytes"
	"crypto/sha256"

	e "github.com/cloudflare/circl/ecc/bls12381"

	"fmt"

	"log"
	"os"
)

type node struct {
	parent                  *node
	childNumb               int
	children                []*node
	ownCompressVectorCommit []byte
	ownVectorCommit         e.G1
	leaf                    bool
	certificate             []byte
	duplicate               bool
	id                      int
	witness witnessStruct
}

type verkleTree struct {
	Root   *node
	leafs  []*node
	fanOut int
	pk     PK
}

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

func BuildTree(certs [][]byte, fanOut int, pk PK) *verkleTree {
	var verk verkleTree
	//var leafs []*node

	uneven := false
	duplicates := fanOut - len(certs)%fanOut

	for len(certs)%fanOut > 0 {
		certs = append(certs, certs[len(certs)-1])
		uneven = true
	}

	leafs := make([]*node, len(certs))

	for i := 0; i < len(certs); i++ {
		byteCet := certs[i]
		//testHash := sha256.Sum256(byteCet)
		leafs[i] = &node{
			certificate: byteCet,
			childNumb:   i % fanOut,
			leaf:        true,
			duplicate:   false,
			id:          i,
		}
	}
	if uneven {
		for i := 1; i < duplicates+1; i++ {
			leafs[len(leafs)-i].duplicate = true
			leafs[len(leafs)-i].childNumb = leafs[len(leafs)-i].id % fanOut
		}
	}

	nextLayer := makeLayer(leafs, fanOut, true, pk)
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer, fanOut, false, pk)
	}

	verk = verkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
		pk:     pk,
	}

	return &verk
}

func makeLayer(nodes []*node, fanOut int, firstLayer bool, pk PK) []*node {

	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate:     nodes[len(nodes)-1].certificate,
			childNumb:       (nodes[len(nodes)-1].id + 1) % fanOut,
			ownVectorCommit: nodes[len(nodes)-1].ownVectorCommit,
			children:        nodes[len(nodes)-1].children,
			leaf:            false,
			duplicate:       true,
			id:              nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)
	}

	nextLayer := make([]*node, len(nodes)/fanOut) // divided with fanout which is length of vectors

	for i := 0; i < len(nodes); {
		childrenList := make([]*node, fanOut)
		//var vectToCommit [][]byte
		vectToCommit := make([][]byte, fanOut)
		if firstLayer {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k] = nodes[i+k].certificate
			}
		} else {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k] = nodes[i+k].ownCompressVectorCommit // this does magic we might need to check it doesn't ruin the list
			}
		}

		polynomial := certVectorToPolynomial(vectToCommit)
		commitment := commit(pk, polynomial)

		nextLayer[i/fanOut] = &node{
			ownVectorCommit:         commitment,
			ownCompressVectorCommit: commitment.BytesCompressed(),
			childNumb:               i % fanOut,
			leaf:                    false,
			children:                childrenList,
			id:                      i / fanOut,
		}
		for _, v := range childrenList {
			v.parent = nextLayer[i/fanOut]
			v.witness = createWitness(pk, polynomial, uint64(v.childNumb))
		}
		i = i + fanOut
	}
	return nextLayer
}

func verifyTree(certs [][]byte, tree verkleTree, pk PK) bool {
	testTree := BuildTree(certs, tree.fanOut, pk)

	return testTree.Root.ownVectorCommit == tree.Root.ownVectorCommit

}

func verifyNode(cert []byte, tree verkleTree, pk PK) bool {
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
	
	for nod.parent != nil {
		witnessIsTrue := verifyWitness(pk, nod.parent.ownVectorCommit ,nod.witness)
		if !witnessIsTrue{
			return false
		}
		nod = nod.parent

	}

	return true
}

func updateLeaf(oldCert []byte, tree verkleTree, newCert []byte) *verkleTree {
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
				//hashList = append(hashList, v.ownVectorCommit)
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
		//nod.parent.ownVectorCommit = sum
		nod = nod.parent
	}
	return &tree
}

func insertLeaf(cert []byte, tree verkleTree) *verkleTree {
	//TODO: Insert a node or delete a node?
	//HOw to do this, what is required??
	return &tree
}

func deleteLeaf(cert []byte, tree verkleTree) *verkleTree {
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
