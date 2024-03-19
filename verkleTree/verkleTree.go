package verkletree

import (
	"bytes"
	"crypto/sha256"

	"time"
	"regexp"
	"slices"
	"strconv"
	"strings"

	e "github.com/cloudflare/circl/ecc/bls12381"

	//	combin "gonum.org/v1/gonum/stat/combin"

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
	polynomial              poly
}

// Membership proof struct containt the neccessary information to verify node belongs to tree.
type membershipProof struct {
	witnesses   []witnessStruct
	commitments []e.G1
}

// Struct representing the verkle-tree
type verkleTree struct {
	Root              *node
	leafs             []*node
	fanOut            int
	pk                PK
	degreeComb        [][][]int
	dividentList      []e.Scalar
	lagrangeBasisList [][]e.Scalar
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
func loadCertificates(input string, amount int) [][]byte {
	var fileArray [][]byte

	for i := 0; i < amount; i++ {
		fileArray = append(fileArray, loadCertificatesFromOneFile(input+"-"+strconv.Itoa(i))...)
	}
	return fileArray
}

func loadCertificatesFromOneFile(input string, amount ...int) [][]byte {

	content, err := os.ReadFile("testCerts/" + input)
	if err != nil {
		panic(err)
	}

	// Convert byte slice to string
	text := string(content)

	// Define regular expression to extract certificates
	certRegex := regexp.MustCompile(`(?s)-----BEGIN CERTIFICATE-----(.*?)-----END CERTIFICATE-----`)

	// Find all matches of certificates
	matches := certRegex.FindAllStringSubmatch(text, -1)

	// Initialize slice to store certificates
	var certificates [][]byte
	if len(amount) == 0 {
		certificates = make([][]byte, len(matches))
	} else {
		certificates = make([][]byte, amount[0])
	}

	// Extract certificates and store them in the slice
	for i, match := range matches {
		if len(amount) != 0 && i == amount[0] {
			return certificates
		}
		certificates[i] = []byte(strings.TrimSpace(match[0]))
	}
	//for i, cert := range certificates {
	//	fmt.Printf("Certificate %d:\n%s\n\n", i+1, cert)
	//}
	return certificates
}

// TODO Taken from combin, make proper cite
// Calculates the unique combination of the integers in the range of 0 to k-1, 0 to k-2, ..., 0.
// Returns the combinations as [][][]int list.
func combinations(n, k int) [][]int {
	combins := Binomial(n, k)
	data := make([][]int, combins)
	if len(data) == 0 {
		return data
	}
	data[0] = make([]int, k)
	for i := range data[0] {
		data[0][i] = i
	}
	next2 := data[0]
	//fmt.Println("len before", len(data[0]))
	//data[0] = data[0][1:]
	//fmt.Println("len after", len(data[0]))
	counter := 1
	for i := 1; i < combins; i++ {
		next := make([]int, k)
		copy(next, next2)
		nextCombination(next, n, k)
		next2 = next
		if slices.Contains(next, 0) {
			next = []int{0}
		} else {
			data[counter] = next
			counter++
		}
		//data[i] = next
	}
	stopIndex := len(data)
	for i := 0; i < len(data); i++ {
		if len(data[i]) == 0 {
			stopIndex = i
			break
		}
	}
	//fmt.Println("len before", len(data))
	//fmt.Println("fær", data)
	data = data[:stopIndex]
	if len(data) > 0 {
		data = data[1:]
	}
	//fmt.Println("efter", data)
	//fmt.Println("len after", len(data))
	return data
}

// TODO Taken from combin, make proper cite
// nextCombination generates the combination after s, overwriting the input value.
func nextCombination(s []int, n, k int) {
	for j := k - 1; j >= 0; j-- {
		if s[j] == n+j-k {
			continue
		}
		s[j]++
		for l := j + 1; l < k; l++ {
			s[l] = s[j] + l - j
		}
		break
	}
}

func Binomial(n, k int) int {
	if n < 0 || k < 0 {
		panic("panic in Binomial")
	}
	if n < k {
		panic("Second panic in Binomial")
	}
	// (n,k) = (n, n-k)
	if k > n/2 {
		k = n - k
	}
	b := 1
	for i := 1; i <= k; i++ {
		b = (n - k + i) * b / i
	}
	return b
}

func combCalculater(fanOut int) [][][]int {

	var degreeComb [][][]int
	for k := fanOut - 1; k > 0; k-- {
		combinationten := combinations(fanOut, k-1)
		//fmt.Println("len of comb to add:", len(combinationten))
		degreeComb = append(degreeComb, combinationten)
	}
	return degreeComb
}

// This function takes the certificates as bytes, the fanout and public key as input.
// Outputs the finished verkle-tree, with the specified fanout.
func BuildTree(certs [][]byte, fanOut int, pk PK) *verkleTree {
	var verk verkleTree

	//Creates a leaf-node for each certificate.
	//fmt.Println("About to create list for nodes")
	leafs := make([]*node, len(certs))
	//fmt.Println("ABout to start for loop to fill leaves")
	for i := 0; i < len(certs); i++ {
		leafs[i] = &node{
			certificate: certs[i],
			childNumb:   i % fanOut,
			duplicate:   false,
			id:          i,
		}
	}

	//Makes the combinations of integers needed to calculate divident and polynomial.
	//fmt.Println("before comb calc")
	degreeComb := combCalculater(fanOut)
	//fmt.Println("After comb calc")
	// define the dividents needed for for calculating the polynomial, these are the same for all polynomial/vectors of the given fanOut size
	dividentList := dividentCalculator(fanOut, degreeComb)
	//start := time.Now()
	//fmt.Println("After div calc")
	lagrangeBasisList := lagrangeBasisCalc(fanOut, degreeComb, dividentList)
	//elapsed := time.Since(start)
	//fmt.Println("Time for langrangeBasis: ", elapsed)

	// call to makeLayer to create next layer in the tree
	start := time.Now()
	
	nextLayer := makeLayer(leafs, fanOut, true, pk, lagrangeBasisList)
	
	// while loop that exits when we are in the root
	for len(nextLayer) > 1 {
		nextLayer = makeLayer(nextLayer, fanOut, false, pk, lagrangeBasisList)
	}
	// Creates the final verkletree struct.
	verk = verkleTree{
		fanOut:            fanOut,
		Root:              nextLayer[0],
		leafs:             leafs,
		pk:                pk,
		lagrangeBasisList: lagrangeBasisList,
	}
	elapsed := time.Since(start)
	fmt.Println("Time elapsed making verkletree: ", elapsed)
	return &verk
}

// Handles the creation of the next layer of the verkle tree. Takes the nodes of the previous layer, the fanout, a bool specifying if it is the first layer and the public key as input.
// Outputs the next layer in the verkle-tree, with size ⌈len(nodes)/fanout⌉. While also adding the witness that each of the layers children belongs to their parents vector commitments.
func makeLayer(nodes []*node, fanOut int, firstLayer bool, pk PK, lagrangeBasisList [][]e.Scalar) []*node {

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
	//var sumTimer int64
	//var sumTimer2 int64

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
		//start := time.Now()
		polynomial := certVectorToPolynomial(vectToCommit, lagrangeBasisList)
		//elapsed := time.Since(start)
		//sumTimer += elapsed.Milliseconds()

		//start = time.Now()
		commitment := commit(pk, polynomial)
		//elapsed = time.Since(start)
		//sumTimer2 += elapsed.Milliseconds()
		//Creates the node with children and vectorcommit.
		nextLayer[i/fanOut] = &node{
			ownVectorCommit:         commitment,
			ownCompressVectorCommit: commitment.BytesCompressed(),
			childNumb:               i % fanOut,
			children:                childrenList,
			id:                      i / fanOut,
			polynomial:              polynomial,
		}
		//Sets the parent in each of the nodes children.
		for _, v := range childrenList {
			v.parent = nextLayer[i/fanOut]
			//v.witness = createWitness(pk, polynomial, uint64(v.childNumb))
		}
		i = i + fanOut
	}
	//fmt.Println("sumTimer for vector to poly", sumTimer)
	//fmt.Println("sumTimer for commit", sumTimer2)

	return nextLayer
}

// This function takes the certificates, verkletree and public key as input. nIt verifies that the verkletree is built using the given certificates.
// Returns true if the tree was correctly built and false if not.
func verifyTree(certs [][]byte, tree verkleTree, pk PK) bool {
	testTree := BuildTree(certs, tree.fanOut, pk)

	return testTree.Root.ownVectorCommit == tree.Root.ownVectorCommit
}

func verifyNode(cert []byte, tree verkleTree) bool {
	mp := createMembershipProof(cert, tree)
	return verifyMembershipProof(mp, tree.pk)
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
	witnessStructEmpty := witnessStruct{}
	for nod.parent != nil {
		if nod.witness == witnessStructEmpty {
			nod.witness = createWitness(tree.pk, nod.parent.polynomial, uint64(nod.childNumb))
		}
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

func dumbUpdateLeaf(tree verkleTree, newCert []byte, oldCert []byte) (verkleTree, bool) {
	var nod *node
	notInList := true
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, oldCert) {
			nod = v
			notInList = false
		}
	}
	if notInList {
		return tree, false
	}
	nod.certificate = newCert
	listlist := make([][]byte, tree.fanOut)
	dumbBool := true
	for nod.parent != nil {
		//childNumber = nod.id % tree.fanOut
		if dumbBool {
			for i, v := range nod.parent.children {
				listlist[i] = v.certificate
			}
			dumbBool = false
		} else {
			for i, v := range nod.parent.children {
				listlist[i] = v.ownCompressVectorCommit
			}
		}

		polyVector := certVectorToPolynomial(listlist, tree.lagrangeBasisList) //TODO what do we do with degComb and divList
		commitment := commit(tree.pk, polyVector)
		nod = nod.parent
		nod.ownVectorCommit = commitment
		nod.ownCompressVectorCommit = commitment.BytesCompressed()
	}

	return tree, true
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

// TODO Not finished, only works for certs%fanout != 0.
func insertLeaf(cert []byte, tree verkleTree) (verkleTree, bool) {
	if len(tree.leafs)%tree.fanOut != 0 {
		finalLeaf := tree.leafs[len(tree.leafs)-1]
		nextSibling := finalLeaf.parent.children[len(tree.leafs)%tree.fanOut]
		nextSibling.certificate = cert
		nextSibling.duplicate = false
		firstLayer := true
		listlist := make([][]byte, tree.fanOut)
		for nextSibling.parent != nil {
			if firstLayer {
				for i, v := range nextSibling.parent.children {
					listlist[i] = v.certificate
				}
				firstLayer = false
			} else {
				for i, v := range nextSibling.parent.children {
					listlist[i] = v.ownCompressVectorCommit
				}
			}

			polyVector := certVectorToPolynomial(listlist, tree.lagrangeBasisList) //TODO what do we do with degComb and divList
			commitment := commit(tree.pk, polyVector)
			nextSibling = nextSibling.parent
			nextSibling.ownVectorCommit = commitment
			nextSibling.ownCompressVectorCommit = commitment.BytesCompressed()
		}
		return tree, true

	}

	return tree, false
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
