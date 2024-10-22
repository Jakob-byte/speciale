package verkletree

import (
	"bytes"
	"encoding/json"

	"slices"
	"sync"

	e "github.com/cloudflare/circl/ecc/bls12381"

	"log"
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
	Witnesses   []witnessStruct
	Commitments []e.G1
}

// Struct representing the verkle-tree
type verkleTree struct {
	Root              *node
	leafs             []*node
	fanOut            int
	pk                PK
	degreeComb        [][][]int
	divisorList       []e.Scalar
	lagrangeBasisList [][]e.Scalar
}

type membershipProofPortable struct {
	Commits [][]byte
	Index   []uint64
	Fx0     [][]byte
	W       [][]byte
}

// Function to call with error to avoid overloading methdods with error if statements
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Creates a json from the witness, and returns it. Logs a fatal error if it fails.
func createJsonOfMembershipProof(mp membershipProof) []byte {
	commits := make([][]byte, len(mp.Commitments))
	for i, v := range mp.Commitments {
		commits[i] = v.BytesCompressed()
	}
	index := make([]uint64, len(mp.Witnesses))
	fx0 := make([][]byte, len(mp.Witnesses))
	w := make([][]byte, len(mp.Witnesses))

	for i, v := range mp.Witnesses {
		index[i] = v.Index
		fx0[i], _ = v.Fx0.MarshalBinary() //TODO fix
		w[i] = v.W.BytesCompressed()
	}
	memProofPort := membershipProofPortable{Commits: commits, Index: index, Fx0: fx0, W: w}
	jSoooon, err := json.Marshal(memProofPort)
	check(err)
	return jSoooon
}

// Retrieves the membership proof from the provided json. Crashes everything otherwise.
func retrieveMembershipProofFromJson(jsonFile []byte) membershipProof {
	var unMarshalled membershipProofPortable
	json.Unmarshal(jsonFile, &unMarshalled)

	commits := make([]e.G1, len(unMarshalled.Commits))
	for i, v := range unMarshalled.Commits {
		commits[i].SetBytes(v)
	}
	length := len(unMarshalled.Index)
	witnesss := make([]witnessStruct, length)

	for i := range length {
		witnesss[i].Fx0.SetBytes(unMarshalled.Fx0[i])
		witnesss[i].Index = unMarshalled.Index[i]
		witnesss[i].W.SetBytes(unMarshalled.W[i])
	}
	return membershipProof{Witnesses: witnesss, Commitments: commits}
}

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
	}
	stopIndex := len(data)
	for i := 0; i < len(data); i++ {
		if len(data[i]) == 0 {
			stopIndex = i
			break
		}
	}
	data = data[:stopIndex]
	if len(data) > 0 {
		data = data[1:]
	}
	return data
}

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
		degreeComb = append(degreeComb, combinationten)
	}
	return degreeComb
}

// This function takes the certificates as bytes, the fanout and public key as input.
// Outputs the finished verkle-tree, with the specified fanout.
func BuildTree(certs [][]byte, fanOut int, pk PK, numThreads ...int) *verkleTree {
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

	//Makes the combinations of integers needed to calculate divisor and polynomial.
	degreeComb := combCalculater(fanOut)

	divisorList := divisorCalculator(fanOut, degreeComb)

	lagrangeBasisList := lagrangeBasisCalc(fanOut, degreeComb, divisorList)

	// call to makeLayer to create next layer in the tree

	//dup nodes
	if len(leafs)%fanOut > 0 {
		leafs = naiveDuplicateNodes(leafs, fanOut)
	}

	//New thread stuff, comment out to make code run -----------------------------------------------------------------------------------------
	if len(numThreads) == 0 {
		numThreads = append(numThreads, 1)
	}

	//setup for starting to build tree setting nextLayer as the starting point, which is the leaf nodes, setting isLeafs to true.
	nextLayer := leafs
	isLeafs := true

	//While loop that keeps making layers of the tree until the length of the nextlayer is >1 which means we are at the root of the tree
	for len(nextLayer) > 1 {

		nodesPerThread := nodesPerThreadCalc(fanOut, len(nextLayer), numThreads[0])

		var nodesForThread []*node
		nextLayer2 := make([][]*node, numThreads[0])
		var mu sync.Mutex
		var wg sync.WaitGroup

		//Collects the nodes for each thread, and starts the process of making the layer with a go routine
		for i := 0; i < len(nextLayer); {
			lastIndex := i + int(nodesPerThread)
			if lastIndex < len(nextLayer) {
				nodesForThread = nextLayer[i:lastIndex]
			} else {
				if len(nextLayer[i:]) < 1 {
					continue
				}
				nodesForThread = nextLayer[i:]
			}
			wg.Add(1)
			go func(index int, nodesToUse []*node, isLeafs2 bool) {
				defer wg.Done()
				makeLayer(nodesToUse, fanOut, isLeafs2, pk, lagrangeBasisList, (index), &nextLayer2, &mu)
			}(i/nodesPerThread, nodesForThread, isLeafs)

			i = i + nodesPerThread
		}
		wg.Wait()

		// collects each threads list of nodes into a single slice of nodes, for the next layer
		nextLayer = []*node{}
		for _, v := range nextLayer2 {
			nextLayer = append(nextLayer, v...)
		}

		if isLeafs {
			isLeafs = false
		}

	}

	//Defines the verkleTree Struct
	verk = verkleTree{
		fanOut:            fanOut,
		Root:              nextLayer[0],
		leafs:             leafs,
		pk:                pk,
		lagrangeBasisList: lagrangeBasisList,
	}

	return &verk
}

func naiveDuplicateNodes(nodes []*node, fanOut int) []*node {
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
	return nodes
}

// Handles the creation of the next layer of the verkle tree. Takes the nodes of the previous layer, the fanout, a bool specifying if it is the first layer and the public key as input.
// Outputs the next layer in the verkle-tree, with size ⌈len(nodes)/fanout⌉. While also adding the witness that each of the layers children belongs to their parents vector commitments.
func makeLayer(nodes []*node, fanOut int, firstLayer bool, pk PK, lagrangeBasisList [][]e.Scalar, index int, nextlayerPointer *[][]*node, mu *sync.Mutex) []*node {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	if len(nodes)%fanOut > 0 {
		nodes = naiveDuplicateNodes(nodes, fanOut)
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
		polynomial := certVectorToPolynomial(vectToCommit, lagrangeBasisList)

		commitment := commit(pk, polynomial)
		//Creates the node with children and vectorcommit.
		nextLayer[i/fanOut] = &node{
			ownVectorCommit:         commitment,
			ownCompressVectorCommit: commitment.BytesCompressed(),
			childNumb:               i % fanOut,
			children:                childrenList,
			id:                      i + nodes[0].id/fanOut,
			polynomial:              polynomial,
		}
		//Sets the parent in each of the nodes children.
		for _, v := range childrenList {
			v.parent = nextLayer[i/fanOut]
		}
		i = i + fanOut
	}
	// Locks the mutex for the nextLayer slice so that the thread can correctly input its nodes to the slice and defers the unlock so it unlocks when finished
	mu.Lock()
	defer mu.Unlock()
	(*nextlayerPointer)[index] = nextLayer
	return nextLayer
}

// This function takes the certificates, verkletree and public key as input. nIt verifies that the verkletree is built using the given certificates.
// Returns true if the tree was correctly built and false if not.
func verifyTree(certs [][]byte, tree verkleTree, pk PK, numThreads int) bool {
	testTree := BuildTree(certs, tree.fanOut, pk, numThreads)

	return testTree.Root.ownVectorCommit == tree.Root.ownVectorCommit
}

// Verifes a specific certificate is in the tree, by first calling createMemberShipProof for the given certificate, and then returns a call to verifyMemberShipProof
func verifyNode(certificate []byte, tree verkleTree) bool {
	mp := createMembershipProof(certificate, tree)
	return verifyMembershipProof(mp, tree.pk)
}

// This function verifies the certificate certificateis commited to in the verkle tree. It takes the certificate, verkle tree and public key as input.
//
//	It returns true if the certificate is in the tree and wrong if it isn't.
func createMembershipProof(certificate []byte, tree verkleTree) membershipProof {
	var nod *node

	notInList := true

	//Retrieves all the certificates from the leaf nodes, to make them searchable with binary search
	certs := make([][]byte, len(tree.leafs))
	for i := range len(tree.leafs) {
		certs[i] = tree.leafs[i].certificate
	}

	//Performs binary search on the leafs, and returns if it found something and what it found.
	n, found := slices.BinarySearchFunc(certs, certificate, func(a, b []byte) int {
		return bytes.Compare(a, b)
	})
	notInList = !found

	var witnessList []witnessStruct
	var commitList []e.G1
	if notInList {
		return membershipProof{}
	}
	nod = tree.leafs[n]
	//Creates the lists required for membership proof
	witnessStructEmpty := witnessStruct{}
	//Calculates the witness up until we see the root
	for nod.parent != nil {
		if nod.witness == witnessStructEmpty {
			nod.witness = createWitness(tree.pk, nod.parent.polynomial, uint64(nod.childNumb))
		}
		witnessList = append(witnessList, nod.witness)
		commitList = append(commitList, nod.parent.ownVectorCommit)
		nod = nod.parent
	}

	return membershipProof{Witnesses: witnessList, Commitments: commitList}
}

// Verifies the membership proof it receives as input with the public key.
// Returns true if the proof is correct, false if it isn't.
func verifyMembershipProof(mp membershipProof, pk PK) bool {
	if len(mp.Witnesses) == 0 {
		return false
	}
	for i := 0; i < len(mp.Witnesses); i++ {
		witnessIsTrue := verifyWitness(pk, mp.Commitments[i], mp.Witnesses[i])
		if !witnessIsTrue {
			return false
		}
	}
	return true
}
