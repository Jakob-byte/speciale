package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"slices"
	"sort"

	"strconv"

	"log"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
)

// struct for representing a node in the tree
type node struct {
	parent      *node
	children    []*node
	ownHash     [32]byte
	certificate []byte
	duplicate   bool
	id          int
}

type witness struct {
	HashList        [][][32]byte
	ChildNumberList []int
}

// Struct representing the merkle-tree
type merkleTree struct {
	Root      *node
	leafs     []*node
	fanOut    int
	pubParams publicParameters
}

type publicParameters struct {
	rootHash [32]byte
	fanOut   int
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
func loadCertificates(input string, amount int, numThreads int) [][]byte {
	var fileArray [][]byte
	files := 1
	stuffToRead := 20000
	if amount > stuffToRead {
		files = int(math.Ceil(float64(amount) / float64(stuffToRead)))
	}
	certThreadList := make([][][]byte, (amount/stuffToRead)+1)

	var mu sync.Mutex
	var wg sync.WaitGroup
	guard := make(chan struct{}, numThreads)

	for i := 0; i < files; i++ {

		if i == files-1 {
			stuffToRead = amount - i*stuffToRead
		}
		wg.Add(1)
		guard <- struct{}{}
		go func(index int, amountToRead int) {
			defer wg.Done()

			loadCertificatesFromOneFile(input+"-"+strconv.Itoa(index)+".crt", index, &certThreadList, &mu, amountToRead)

			<-guard
		}(i, stuffToRead)
	}
	wg.Wait()

	for _, v := range certThreadList {

		if len(v) > 0 {
			fileArray = append(fileArray, v...)
		}
	}
	//Sorts the certificates
	sort.Slice(fileArray, func(i, j int) bool { return (bytes.Compare(fileArray[i], fileArray[j]) == -1) })
	return fileArray
}

// Loads the specified amount of certificates.
func loadCertificatesFromOneFile(input string, index int, listPoint *[][][]byte, mu *sync.Mutex, amount ...int) {
	content, err := os.ReadFile("../testCerts/" + input)
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
			break
		}
		certificates[i] = []byte(strings.TrimSpace(match[0]))
	}

	mu.Lock()
	defer mu.Unlock()
	(*listPoint)[index] = certificates
}

// Creates a json from the witness, and returns it. Logs a fatal error if it fails.
func genJsonWitness(wit witness) []byte {
	json, err := json.Marshal(wit)
	check(err)
	return json

}

// Retrieves the witness from the json.
func getWitnessFromJson(jsonWit []byte) witness {
	var unMarshalled witness
	json.Unmarshal(jsonWit, &unMarshalled)
	return unMarshalled
}

// Calculates the amount of nodes that should be given to each thread
func nodesPerThreadCalc(fanOut, lenNextLayer, numThreads int) int {
	NodePerThreadcalc := float64(lenNextLayer) / float64(fanOut)
	NodePerThreadcalc = math.Ceil(NodePerThreadcalc/float64(numThreads)) * float64(fanOut)
	nodesPerThread := int(NodePerThreadcalc)
	return nodesPerThread
}

// This function builds the merkle tree from the provided certificates with the desired fan-out.
func BuildTree(certs [][]byte, fanOut int, numThreads ...int) *merkleTree {
	var merk merkleTree

	leafs := make([]*node, len(certs))

	//build the leaf nodes of the tree
	for i := 0; i < len(certs); i++ {
		leafs[i] = &node{
			certificate: certs[i],
			ownHash:     sha256.Sum256(certs[i]),
			duplicate:   false,
			id:          i,
		}

	}
	if len(leafs)%fanOut > 0 {
		leafs = duplicateNodes(leafs, fanOut)
	}

	if len(numThreads) == 0 {
		numThreads = append(numThreads, 1)
	}

	nextLayer := leafs
	for len(nextLayer) > 1 {
		nodesPerThread := nodesPerThreadCalc(fanOut, len(nextLayer), numThreads[0])
		var nodesForThread []*node
		nextLayer2 := make([][]*node, numThreads[0])
		var mu sync.Mutex
		var wg sync.WaitGroup
		for i := 0; i < len(nextLayer); {
			lastIndex := i + nodesPerThread
			if lastIndex < len(nextLayer) {
				nodesForThread = nextLayer[i:lastIndex]
			} else {
				if len(nextLayer[i:]) < 1 {
					continue
				}
				nodesForThread = nextLayer[i:]
			}
			wg.Add(1)
			go func(index int, nodesToUse []*node) {
				defer wg.Done()
				makeLayer(nodesToUse, fanOut, index, &nextLayer2, &mu)
			}(i/nodesPerThread, nodesForThread)

			i = i + nodesPerThread
		}
		wg.Wait()

		nextLayer = []*node{}

		for _, v := range nextLayer2 {
			nextLayer = append(nextLayer, v...)
		}

	}

	// define the merkletree struct
	merk = merkleTree{
		fanOut:    fanOut,
		Root:      nextLayer[0],
		leafs:     leafs,
		pubParams: publicParameters{fanOut: fanOut, rootHash: nextLayer[0].ownHash},
	}
	return &merk
}

// This function duplicates the nodes.
func duplicateNodes(nodes []*node, fanOut int) []*node {
	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate: nodes[len(nodes)-1].certificate,
			ownHash:     nodes[len(nodes)-1].ownHash,
			children:    nodes[len(nodes)-1].children,
			duplicate:   true,
			id:          1 + nodes[len(nodes)-1].id,
		}
		nodes = append(nodes, appendNode)
	}

	return nodes
}

// This function handles the creation of each layer.
func makeLayer(nodes []*node, fanOut int, index int, nextlayerPointer *[][]*node, mu *sync.Mutex) []*node {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	if len(nodes)%fanOut > 0 {
		nodes = duplicateNodes(nodes, fanOut)
	}

	nextLayer := make([]*node, len(nodes)/fanOut) // divided with fanout

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
		//Creates the node with children and vectorcommit.
		nextLayer[i/fanOut] = &node{
			ownHash:  sha256.Sum256(allChildrenHashes),
			children: childrenList,
			id:       i/fanOut + (len(nodes)/fanOut)*index,
		}

		////Sets the parent for each of the nodes in the now previous layer.
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

// verifies that the input tree is made of the input certs, by rebuilding the tree from the certificates and comparing the roothash
func verifyTree(certs [][]byte, tree merkleTree) bool {
	testTree := BuildTree(certs, tree.fanOut)

	return testTree.Root.ownHash == tree.Root.ownHash

}

// Verifies a node is in the tree by creating witness and veryifing the witness
func verifyNode(certificate []byte, tree merkleTree) bool {
	witness := createWitness(certificate, tree)
	return verifyWitness(certificate, witness, tree.pubParams)
}

// Creates the witness for a given certificate, which is a witness struct consisting of a list of hashes and a childNumberList to know how to combine the hashes
func createWitness(certificate []byte, tree merkleTree) witness {
	var node *node
	notInList := true

	//Retrieves all the certificates from the leaf nodes
	certs := make([][]byte, len(tree.leafs))
	for i := range len(tree.leafs) {
		certs[i] = tree.leafs[i].certificate
	}

	//Performs binary search on the leafs, and returns if it found something and what it found.
	n, found := slices.BinarySearchFunc(certs, certificate, func(a, b []byte) int {
		return bytes.Compare(a, b)
	})
	notInList = !found

	if notInList {
		return witness{}
	}
	node = tree.leafs[n]
	var hashList [][][32]byte

	var childNumberList []int

	var counter int
	for node.parent != nil {
		tempHashList := make([][32]byte, tree.fanOut-1)

		childNumberList = append(childNumberList, node.id%tree.fanOut)
		counter = 0
		for _, v := range node.parent.children {
			if node.id != v.id {

				tempHashList[counter] = v.ownHash
				counter++
			}
		}

		hashList = append(hashList, tempHashList)
		node = node.parent
	}

	witness := witness{
		HashList:        hashList,
		ChildNumberList: childNumberList,
	}
	return witness
}

// verifies the witness given a initial certificate
func verifyWitness(certificate []byte, witness witness, pubParams publicParameters) bool {
	sum := sha256.Sum256(certificate)
	for i := 0; i < len(witness.HashList); i++ {
		var byteToHash []byte
		for j, v := range witness.HashList[i] {

			if witness.ChildNumberList[i] == j {
				byteToHash = append(byteToHash, sum[:]...)
			}
			byteToHash = append(byteToHash, v[:]...)
		}
		if witness.ChildNumberList[i] == pubParams.fanOut-1 {
			byteToHash = append(byteToHash, sum[:]...)
		}
		sum = sha256.Sum256(byteToHash)
	}

	return sum == pubParams.rootHash

}

func main() {

}
