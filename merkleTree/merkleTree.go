package main

import (
	"bytes"
	"crypto/sha256"

	//"errors"
	"fmt"
	"strconv"

	//"hash"
	"sync"

	"log"
	"math"
	"os"
	"regexp"
	"strings"
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

type witness struct {
	hashList        [][][32]byte
	childNumberList []int
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
}

// function to load certificates given, input which is the directory and amount represented as a list of ints,
// where [0} is the amount of certificates to load from said directory.
// returns a [][]byte list/array of files
func loadCertificates(input string, amount int) [][]byte {
	var fileArray [][]byte
	files := 1
	stuffToRead := 20000
	if amount > stuffToRead {
		files = int(math.Ceil(float64(amount) / float64(stuffToRead)))
	}
	for i := 0; i < files; i++ {
		if i == files-1 {
			stuffToRead = amount - i*stuffToRead
		}
		fileArray = append(fileArray, loadCertificatesFromOneFile(input+"-"+strconv.Itoa(i)+".crt", stuffToRead)...)
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

func BuildTree(certs [][]byte, fanOut int, numThreads ...int) *merkleTree {
	var merk merkleTree

	nodes := make([]*node, len(certs))

	//build the leaf nodes of the tree
	for i := 0; i < len(certs); i++ {
		nodes[i] = &node{
			certificate: certs[i],
			childNumb:   i % fanOut,
			ownHash:     sha256.Sum256(certs[i]),
			duplicate:   false,
			id:          i,
		}

	}
	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate: nodes[len(nodes)-1].certificate,
			childNumb:   (nodes[len(nodes)-1].id + 1) % fanOut,
			ownHash:     nodes[len(nodes)-1].ownHash,
			duplicate:   true,
			id:          nodes[len(nodes)-1].id + 1,
		}
		nodes = append(nodes, appendNode)
	}

	//function call to make the next layer

	if len(numThreads) == 0 {
		numThreads = append(numThreads, 1)
	}

	nextLayer := nodes
	for len(nextLayer) > 1 {
		NodePerThreadcalc := float64(len(nextLayer)) / float64(fanOut)
		// 1000/3 = 333,333333
		NodePerThreadcalc = math.Ceil(NodePerThreadcalc/float64(numThreads[0])) * float64(fanOut)
		//(3333,3333 / 4) * 3 => 84 * 3 =  252
		nodesPerThread := int(NodePerThreadcalc)
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

		//fmt.Println("layer done -----------------------------------------------")

		nextLayer = []*node{}

		for _, v := range nextLayer2 {
			nextLayer = append(nextLayer, v...)
		}
		//for _, v := range nextLayer{
		//	fmt.Println("Look an important id:", v.id)
		//}
		//		fmt.Println("layer START -----------------------------------------------")

	}

	//nextLayer := makeLayer(leafs, fanOut)
	//// Checks if nextlayer it the root by checking the length, if not root call nextlayer again
	//for len(nextLayer) > 1 {
	//	nextLayer = makeLayer(nextLayer, fanOut)
	//}

	// define the merkletree struct
	merk = merkleTree{
		fanOut: fanOut,
		Root:   nextLayer[0],
		leafs:  nodes,
	}
	return &merk
}

func makeLayer(nodes []*node, fanOut int, index int, nextlayerPointer *[][]*node, mu *sync.Mutex) []*node {

	//makes the tree balanced according to the fanout, by duplicating the last node until it is balanced
	// hvis forskellige threads gør det her?? så er det jo forskellige id i sidste? eller samme
	for len(nodes)%fanOut > 0 {
		appendNode := &node{
			certificate: nodes[len(nodes)-1].certificate,
			childNumb:   (nodes[len(nodes)-1].id + 1) % fanOut,
			ownHash:     nodes[len(nodes)-1].ownHash,
			children:    nodes[len(nodes)-1].children,
			duplicate:   true,
			id:          1 + nodes[len(nodes)-1].id,
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
		//for _, a := range childrenList {
		//	fmt.Println("THese id's are wrong I hope", a.id)
		//}
		//Creates the hash of the children of the node.

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
	//fmt.Println("Index:", index)
	//fmt.Println("length of next layer", len(*nextlayerPointer))
	(*nextlayerPointer)[index] = nextLayer

	return nextLayer

}

// verifies that the input tree is made of the input certs, by rebuilding the tree from the certificates and comparing the roothash
func verifyTree(certs [][]byte, tree merkleTree) bool {
	testTree := BuildTree(certs, tree.fanOut)

	return testTree.Root.ownHash == tree.Root.ownHash

}

// Verifies a node is in the tree by creating witness and veryifing the witness
func verifyNode(cert []byte, tree merkleTree) bool {
	witness := createWitness(cert, tree)
	//fmt.Println("witness hashList: ", witness.hashList)
	return verifyWitness(cert, witness, tree)
}

// Creates the witness for a given certificate, which is a witness struct consisting of a list of hashes and a childNumberList to know how to combine the hashes
func createWitness(cert []byte, tree merkleTree) witness {
	var nod *node
	//fanOut := tree.fanOut
	notInList := true
	for _, v := range tree.leafs {
		if bytes.Equal(v.certificate, cert) {
			nod = v
			notInList = false
		}
	}
	if notInList {
		return witness{}
	}

	var hashList [][][32]byte

	var childNumberList []int

	var counter int
	for nod.parent != nil {
		hashList0 := make([][32]byte, tree.fanOut-1)

		childNumberList = append(childNumberList, nod.id%tree.fanOut)
		counter = 0
		for _, v := range nod.parent.children {
			if nod.id != v.id {

				hashList0[counter] = v.ownHash
				counter++
			}
		}
		//fmt.Printf("NONWORKINGLIST: %T\n",(hashList0))
		//fmt.Printf("workingLIST: %T\n",(hashList0working))

		//fmt.Println("HashList0, should be equal to fanout", len(hashList0))

		hashList = append(hashList, hashList0)
		nod = nod.parent
	}

	witness := witness{
		hashList:        hashList,
		childNumberList: childNumberList,
	}
	return witness
}

// if we wanna be cool we can fix stuff so we don't need to give the certificate to this function! by putting it in the hashlist somehow
func verifyWitness(cert []byte, witness witness, tree merkleTree) bool {
	sum := sha256.Sum256(cert)
	for i := 0; i < len(witness.hashList); i++ {
		var byteToHash []byte
		for j, v := range witness.hashList[i] {

			if witness.childNumberList[i] == j {
				byteToHash = append(byteToHash, sum[:]...)
			}
			byteToHash = append(byteToHash, v[:]...)
		}
		if witness.childNumberList[i] == tree.fanOut-1 {
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

		// Create the thing we want to hash
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
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	merkTree := BuildTree(certArray, 2)
	fmt.Println("Verify tree works for correct tree", verifyTree(certArray, *merkTree))
	fmt.Println("Verify node works for correct node", verifyNode(certArray[5], *merkTree))
	updatedTree := updateLeaf(certArray[5], *merkTree, certArray[3])
	fmt.Println("We managed to overwrite a certificate", !verifyNode(certArray[5], *updatedTree))
	insertLeaf(certArray[5], *merkTree)
	deleteLeaf(certArray[5], *merkTree)

	fmt.Println("Succes")
}
