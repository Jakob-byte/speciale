package verkletree

import (
	"bytes"
	"encoding/json"

	//"fmt"

	"slices"
	"sync"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

// struct representing the nodes in the verkletree

type optimizedNode struct {
	parent                  *optimizedNode
	childNumb               int
	children                []*optimizedNode
	ownCompressVectorCommit []byte
	ownVectorCommit         e.G1
	certificate             []byte
	duplicate               bool
	//id                      int
	witness optimizedWitnessStruct
}

// Membership proof struct containt the neccessary information to verify node belongs to tree.
type optimizedMembershipProof struct {
	Witnesses   []optimizedWitnessStruct
	Commitments []e.G1
}

// Struct representing the verkle-tree
type optimizedVerkleTree struct {
	Root   *optimizedNode
	leafs  []*optimizedNode
	fanOut int
	pk     pubParams
}

// Creates a json from the witness, and returns it. Logs a fatal error if it fails.
func optimizedCreateJsonOfMembershipProof(mp optimizedMembershipProof) []byte {
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
	mpJson, err := json.Marshal(memProofPort)
	check(err)
	return mpJson
}

// Retrieves the membership proof from the provided json. Crashes everything otherwise.
func optimizedRetrieveMembershipProofFromJson(jsonFile []byte) optimizedMembershipProof {
	var unMarshalled membershipProofPortable
	json.Unmarshal(jsonFile, &unMarshalled)

	commits := make([]e.G1, len(unMarshalled.Commits))
	for i, v := range unMarshalled.Commits {
		commits[i].SetBytes(v)
	}
	length := len(unMarshalled.Index)
	witnesss := make([]optimizedWitnessStruct, length)

	for i := range length {
		witnesss[i].Fx0.SetBytes(unMarshalled.Fx0[i])
		witnesss[i].Index = unMarshalled.Index[i]
		witnesss[i].W.SetBytes(unMarshalled.W[i])
	}
	//fmt.Println("witnesssss", witnesss)
	//fmt.Println("Comits", commits)
	return optimizedMembershipProof{Witnesses: witnesss, Commitments: commits}
}

// This function takes the certificates as bytes, the fanout and public key as input.
// Outputs the finished verkle-tree, with the specified fanout.
func optimizedBuildTree(certs [][]byte, fanOut int, pk pubParams, includeWitnesses bool, numThreads ...int) *optimizedVerkleTree {
	//fmt.Println("BuildTree called with fanout", fanOut)
	var verk optimizedVerkleTree

	//Creates a leaf-node for each certificate.
	leafs := make([]*optimizedNode, len(certs))
	for i := 0; i < len(certs); i++ {
		leafs[i] = &optimizedNode{
			certificate: certs[i],
			childNumb:   i % fanOut,
			duplicate:   false,
			//id:          i,
		}
	}

	//dup nodes
	if len(leafs)%fanOut > 0 {
		leafs = optimizedDuplicateNodes(leafs, fanOut)
	}

	//If numThreads not provided, set it to 1.
	if len(numThreads) == 0 {
		numThreads = append(numThreads, 1)
	}

	//setup for starting to build tree setting nextLayer as the starting point, which is the leaf nodes, setting isLeafs to true.
	nextLayer := leafs
	isLeafs := true

	//While loop that keeps making layers of the tree until the length of the nextlayer is >1 which means we are at the root of the tree
	for len(nextLayer) > 1 {

		// Calculates how many nodes each thread will be assigned and makes a list for the threads to save their output in
		nodesPerThread := nodesPerThreadCalc(fanOut, len(nextLayer), numThreads[0])
		var nodesForThread []*optimizedNode
		nextLayer2 := make([][]*optimizedNode, numThreads[0])
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
			go func(index int, nodesToUse []*optimizedNode, isLeafs2 bool) {
				defer wg.Done()
				optimizedMakeLayer(nodesToUse, fanOut, isLeafs2, pk, index, &nextLayer2, &mu, includeWitnesses)
			}(i/nodesPerThread, nodesForThread, isLeafs)

			i = i + nodesPerThread
		}

		wg.Wait()

		// collects each threads list of nodes into a single slice of nodes, for the next layer
		nextLayer = []*optimizedNode{}
		for _, v := range nextLayer2 {
			nextLayer = append(nextLayer, v...)
		}

		if isLeafs {
			isLeafs = false
		}

	}

	//Defines the verkleTree Struct
	verk = optimizedVerkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  leafs,
		pk:     pk,
	}

	return &verk
}

func optimizedDuplicateNodes(nodes []*optimizedNode, fanOut int) []*optimizedNode {
	for len(nodes)%fanOut > 0 {
		appendNode := &optimizedNode{
			certificate:     nodes[len(nodes)-1].certificate,
			childNumb:       (nodes[len(nodes)-1].childNumb + 1) % fanOut,
			ownVectorCommit: nodes[len(nodes)-1].ownVectorCommit,
			children:        nodes[len(nodes)-1].children,
			duplicate:       true,
			//id:              nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)
	}
	return nodes
}

// Handles the creation of the next layer of the verkle tree. Takes the nodes of the previous layer, the fanout, a bool specifying if it is the first layer and the public key as input.
// Outputs the next layer in the verkle-tree, with size ⌈len(nodes)/fanout⌉. While also adding the witness that each of the layers children belongs to their parents vector commitments.
func optimizedMakeLayer(nodes []*optimizedNode, fanOut int, firstLayer bool, pk pubParams, index int, nextlayerPointer *[][]*optimizedNode, mu *sync.Mutex, includeWitnesses bool) []*optimizedNode {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	if len(nodes)%fanOut > 0 {
		nodes = optimizedDuplicateNodes(nodes, fanOut)
	}
	//Creates the slice for the next layer, which is len(nodes)/fanOut.
	nextLayer := make([]*optimizedNode, len(nodes)/fanOut) // divided with fanout which is length of vectors

	//The for loop which creates the next layer by create the vector commit for each of the new nodes.
	//And adding the corresponding children to each of their parents in the tree.
	//var sumTimer int64
	//var sumTimer2 int64

	for i := 0; i < len(nodes); {
		//The loop starts by finding the children for the current node in the 'nextlayer'

		childrenList := make([]*optimizedNode, fanOut)
		vectToCommit := make([]e.Scalar, fanOut)
		if firstLayer {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k].SetBytes(nodes[i+k].certificate)
			}
		} else {
			for k := 0; k < fanOut; k++ {
				childrenList[k] = nodes[i+k]
				vectToCommit[k].SetBytes(nodes[i+k].ownCompressVectorCommit)
			}
		}
		//Creates the vectorcommit to the children of the node.
		//start := time.Now()
		//elapsed := time.Since(start)
		//sumTimer += elapsed.Milliseconds()
		//start = time.Now()
		commitment := optimizedCommit(pk, vectToCommit)
		//elapsed = time.Since(start)
		//sumTimer2 += elapsed.Milliseconds()
		//Creates the node with children and vectorcommit.

		// TODO SHOULD WE TRY TO PROOF GEN WHILE BUILDING TREE?? TO SEE RUNTIME?
		nextLayer[i/fanOut] = &optimizedNode{

			ownVectorCommit:         commitment,
			ownCompressVectorCommit: commitment.BytesCompressed(),
			childNumb:               i % fanOut,
			children:                childrenList,
			//id:                      i + nodes[0].id/fanOut,
		}
		//Sets the parent in each of the nodes children.
		for j, v := range childrenList {
			if includeWitnesses {
				v.witness = optimizedWitnessStruct{W: optimizedProveGen(pk, vectToCommit, i%fanOut),
					Index: uint64(i % fanOut),
					Fx0:   vectToCommit[j]}
			}
			v.parent = nextLayer[i/fanOut]

		}
		i = i + fanOut
	}
	//fmt.Println("sumTimer for vector to poly", sumTimer)
	//fmt.Println("sumTimer for commit", sumTimer2)
	// Locks the mutex for the nextLayer slice so that the thread can correctly input its nodes to the slice and defers the unlock so it unlocks when finished
	mu.Lock()
	defer mu.Unlock()
	//fmt.Println("Index:", index)
	//fmt.Println("length of next layer", len(*nextlayerPointer))
	(*nextlayerPointer)[index] = nextLayer
	return nextLayer
}

// This function takes the certificates, verkletree and public key as input. nIt verifies that the verkletree is built using the given certificates.
// Returns true if the tree was correctly built and false if not.
func optimizedVerifyTree(certs [][]byte, tree optimizedVerkleTree, pk pubParams, numThreads int) bool {
	testTree := optimizedBuildTree(certs, tree.fanOut, pk, false, numThreads)

	return testTree.Root.ownVectorCommit == tree.Root.ownVectorCommit
}

// Verifes a specific certificate is in the tree, by first calling createMemberShipProof for the given certificate, and then returns a call to verifyMemberShipProof
func optimizedVerifyNode(cert []byte, tree optimizedVerkleTree) bool {
	mp := optimizedCreateMembershipProof(cert, tree)
	return optimizedVerifyMembershipProof(mp, tree.pk)
}

// This function verifies the certificate cert is commited to in the verkle tree. It takes the certificate, verkle tree and public key as input.
//
//	It returns true if the certificate is in the tree and wrong if it isn't.
func optimizedCreateMembershipProof(cert []byte, tree optimizedVerkleTree) optimizedMembershipProof {
	var nod *optimizedNode

	notInList := true
	//Finds the node which has the certificate. If it doesn't exist we return false.
	//for _, v := range tree.leafs {
	//	if bytes.Equal(v.certificate, cert) {
	//		nod = v
	//		notInList = false
	//	}
	//}

	//Retrieves all the certificates from the leaf nodes, to make them searchable with binary search
	certs := make([][]byte, len(tree.leafs))
	for i := range len(tree.leafs) {
		certs[i] = tree.leafs[i].certificate
	}

	//Performs binary search on the leafs, and returns if it found something and what it found.
	n, found := slices.BinarySearchFunc(certs, cert, func(a, b []byte) int {
		return bytes.Compare(a, b)
	})
	notInList = !found

	var witnessList []optimizedWitnessStruct
	var commitList []e.G1
	if notInList {
		return optimizedMembershipProof{}
	}
	nod = tree.leafs[n]
	//Creates the lists required for membership proof
	witnessStructEmpty := optimizedWitnessStruct{}
	//Calculates the witness up until we see the root
	childCommits := make([]e.Scalar, tree.fanOut)
	isLeaf := true
	for nod.parent != nil {

		if nod.witness == witnessStructEmpty {
			if isLeaf {
				for i := range childCommits {
					childCommits[i].SetBytes(nod.parent.children[i].certificate)
				}
				isLeaf = false
			} else {
				for i := range childCommits {
					childCommits[i].SetBytes(nod.parent.children[i].ownCompressVectorCommit)
				}
			}

			nod.witness = optimizedWitnessStruct{W: optimizedProveGen(tree.pk, childCommits, nod.childNumb),
				Index: uint64(nod.childNumb),
				Fx0:   childCommits[nod.childNumb]}
		}
		witnessList = append(witnessList, nod.witness)
		commitList = append(commitList, nod.parent.ownVectorCommit)
		nod = nod.parent
	}

	return optimizedMembershipProof{Witnesses: witnessList, Commitments: commitList}
}

// Verifies the membership proof it receives as input with the public key.
// Returns true if the proof is correct, false if it isn't.
func optimizedVerifyMembershipProof(mp optimizedMembershipProof, pk pubParams) bool {
	if len(mp.Witnesses) == 0 {
		return false
	}
	for i := 0; i < len(mp.Witnesses); i++ {
		witnessIsTrue := optimizedVerify(pk, mp.Commitments[i], mp.Witnesses[i].W, mp.Witnesses[i].Fx0, int(mp.Witnesses[i].Index))
		if !witnessIsTrue {
			return witnessIsTrue //refactored this to witnessIsTrue instead of false
		}
	}
	return true
}
