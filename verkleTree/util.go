package verkletree

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

// Calculates and returns x^n
func powerSimple(x e.Scalar, n int) e.Scalar {
	var result e.Scalar
	result.SetOne()
	for j := 0; j < n; j++ {
		result.Mul(&result, &x)
	}
	fmt.Println("Hej")
	return result
}

// This function loads a single certificate and returns it.
func loadOneCert(filePath string) []byte {
	f, err := os.ReadFile(filePath)
	check(err)
	return f
}

// function to load certificates given, input which is the directory and amount represented as a list of ints,
// where [0] is the amount of certificates to load from said directory.
// returns a [][]byte list/array of files
func loadCertificates(input string, amount int) [][]byte {
	var fileArray [][]byte
	files := 1
	stuffToRead := 20000
	if amount > stuffToRead {
		files = int(math.Ceil(float64(amount) / float64(stuffToRead)))
	}

	certThreadList := make([][][]byte, (amount/stuffToRead)+1)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < files; i++ {
		if i == files-1 {
			stuffToRead = amount - i*stuffToRead
		}
		wg.Add(1)

		go func(index int, amountToRead int) {
			defer wg.Done()
			loadCertificatesFromOneFile(input+"-"+strconv.Itoa(index)+".crt", index, &certThreadList, &mu, amountToRead)
		}(i, stuffToRead)
	}
	wg.Wait()

	for _, v := range certThreadList {

		if len(v) > 0 {
			fileArray = append(fileArray, v...)
		}
	}
	//Sorts the certificates
	sort.Slice(fileArray, func(i, j int) bool { return bytes.Compare(fileArray[i], fileArray[j]) == -1 })
	fmt.Println("length of filearray: ", len(fileArray), " capacity of file Array: ", cap(fileArray))
	return fileArray
}

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
	//for i, cert := range certificates {
	//	fmt.Printf("Certificate %d:\n%s\n\n", i+1, cert)
	//}
	mu.Lock()
	defer mu.Unlock()
	(*listPoint)[index] = certificates
}

func nodesPerThreadCalc(fanOut, lenNextLayer, numThreads int) int {
	NodePerThreadcalc := float64(lenNextLayer) / float64(fanOut)
	NodePerThreadcalc = math.Ceil(NodePerThreadcalc/float64(numThreads)) * float64(fanOut)
	nodesPerThread := int(NodePerThreadcalc)
	return nodesPerThread
}
