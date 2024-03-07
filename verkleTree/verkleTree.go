package verkletree

import (
	"bytes"
	"crypto/sha256"

	e "github.com/cloudflare/circl/ecc/bls12381"
	combin "gonum.org/v1/gonum/stat/combin"

	"fmt"
	"log"
	"os"
)

// struct for representing a node in the tree
type node struct {
	parent                  *node
	childNumb               int
	children                []*node
	ownCompressVectorCommit []byte
	ownVectorCommit         e.G1
	certificate             []byte
	duplicate               bool
	id                      int
	witness                 witnessStruct
}

// Membership proof struct containt the neccessary information to verify node belongs to tree.
type membershipProof struct {
	witnesses   []witnessStruct
	commitments []e.G1
}

// Struct representing the verkle-tree
type verkleTree struct {
	Root   *node
	leafs  []*node
	fanOut int
	pk     PK
}

// Function to call with error to avoid overloading methdods with error if statements
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

// Calculates the unique combination of the integers in the range of 0 to k-1, 0 to k-2, ..., 0.
// Returns the combinations as [][][]int list.
func combCalculater(fanOut int) [][][]int {
	var degreeComb [][][]int
	for k := fanOut - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(fanOut, k-1))
	}
	return degreeComb
}

// This function takes the certificates as bytes, the fanout and public key as input.
// Outputs the finished verkle-tree, with the specified fanout.
func BuildTree(certs [][]byte, fanOut int, pk PK) *verkleTree {
	var verk verkleTree

	//Creates a leaf-node for each certificate.
	leafs := make([]*node, len(certs))
	for i := 0; i < len(certs); i++ {
		leafs[i] = &node{
			certificate: certs[i],
			childNumb:   i % fanOut,
			duplicate:   false,
			id:          i,
		}
	}

	//Makes the combinations of integers needed to calculate divident and polynomial.
	degreeComb := combCalculater(fanOut)
	// define the dividents needed for for calculating the polynomial, these are the same for all polynomial/vectors of the given fanOut size
	dividentList := dividentCalculator(fanOut, degreeComb)

	// call to makeLayer to create next layer in the tree
	nextLayer := makeLayer(leafs, fanOut, true, pk, degreeComb, dividentList)
	// while loop that exits when we are in the root
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer, fanOut, false, pk, degreeComb, dividentList)
	}
	// Creates the final verkletree struct.
	verk = verkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
		pk:     pk,
	}

	return &verk
}

// Handles the creation of the next layer of the verkle tree. Takes the nodes of the previous layer, the fanout, a bool specifying if it is the first layer and the public key as input.
// Outputs the next layer in the verkle-tree, with size ⌈len(nodes)/fanout⌉. While also adding the witness that each of the layers children belongs to their parents vector commitments.
func makeLayer(nodes []*node, fanOut int, firstLayer bool, pk PK, degreeComb [][][]int, dividentList []e.Scalar) []*node {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate:     nodes[len(nodes)-1].certificate,
			childNumb:       (nodes[len(nodes)-1].id + 1) % fanOut,
			ownVectorCommit: nodes[len(nodes)-1].ownVectorCommit,
			children:        nodes[len(nodes)-1].children,
			duplicate:       true,
			id:              nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)
	}
	//Creates the slice for the next layer, which is len(nodes)/fanOut.
	nextLayer := make([]*node, len(nodes)/fanOut) // divided with fanout which is length of vectors

	//The for loop which creates the next layer by create the vector commit for each of the new nodes.
	//And adding the corresponding children to each of their parents in the tree.
	for i := 0; i < len(nodes); {
		//The loop starts by finding the children for the current node in the 'nextlayer'
		childrenList := make([]*node, fanOut)
		vectToCommit := make([][]byte, fanOut)
		if firstLayer {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k] = nodes[i+k].certificate
			}
		} else {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k] = nodes[i+k].ownCompressVectorCommit
			}
		}
		//Creates the vectorcommit to the children of the node.
		polynomial := certVectorToPolynomial(vectToCommit, degreeComb, dividentList)

		commitment := commit(pk, polynomial)
		//Creates the node with children and vectorcommit.
		nextLayer[i/fanOut] = &node{
			ownVectorCommit:         commitment,
			ownCompressVectorCommit: commitment.BytesCompressed(),
			childNumb:               i % fanOut,
			children:                childrenList,
			id:                      i / fanOut,
		}
		//Sets the witness in each of the nodes children.
		for _, v := range childrenList {
			v.parent = nextLayer[i/fanOut]
			v.witness = createWitness(pk, polynomial, uint64(v.childNumb))
		}
		i = i + fanOut
	}
	return nextLayer
}

// This function takes the certificates, verkletree and public key as input. nIt verifies that the verkletree is built using the given certificates.
// Returns true if the tree was correctly built and false if not.
func verifyTree(certs [][]byte, tree verkleTree, pk PK) bool {
	testTree := BuildTree(certs, tree.fanOut, pk)

	return testTree.Root.ownVectorCommit == tree.Root.ownVectorCommit
}

// This function verifies the certificate cert is commited to in the verkle tree. It takes the certificate, verkle tree and public key as input.
//
//	It returns true if the certificate is in the tree and wrong if it isn't.
func createMembershipProof(cert []byte, tree verkleTree) membershipProof {
	var nod *node

	notInList := true
	//Finds the node which has the certificate. If it doesn't exist we return false.
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, cert) {
			nod = v
			notInList = false
		}
	}

	var witnessList []witnessStruct
	var commitList []e.G1
	if notInList {
		return membershipProof{}
	}
	//Creates the lists required for membership proof
	for nod.parent != nil {
		witnessList = append(witnessList, nod.witness)
		commitList = append(commitList, nod.parent.ownVectorCommit)
		nod = nod.parent
	}

	return membershipProof{witnesses: witnessList, commitments: commitList}
}

// Verifies the membership proof it receives as input with the public key.
// Returns true if the proof is correct, false if it isn't.
func verifyMembershipProof(mp membershipProof, pk PK) bool {
	for i := 0; i < len(mp.witnesses); i++ {
		witnessIsTrue := verifyWitness(pk, mp.commitments[i], mp.witnesses[i])
		if !witnessIsTrue {
			return false
		}
	}
	return true

}

// This function is NOT finished
// This function updates a leaf in the tree. It takes the old certificate it replaces, the tree and a new certificate to replace the old with as input.
// Returns the updated tree if the old certificates was in the tree. If it wasn't it just returns the inputted tree.
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

// Not finished
func insertLeaf(cert []byte, tree verkleTree) *verkleTree {
	//TODO: Insert a node or delete a node?
	//HOw to do this, what is required??
	return &tree
}

// Not finished
func deleteLeaf(cert []byte, tree verkleTree) *verkleTree {
	//TODO: Insert a node or delete a node?
	//How to do this, what is required??
	//Diego said this is not required and will be done when rebuilding or somewhere else by the CA
	fmt.Println("Sike! We cannot delete stuff.")
	return &tree
}

// This function loads a single certificate and returns it.
func loadOneCert(filePath string) []byte {
	f, err := os.ReadFile(filePath)
	check(err)
	return f
}
