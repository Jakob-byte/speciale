package verkletree

import (
	"fmt"
	"testing"
	"time"
)

var roottestCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFIle20000", 10000),
}

func TestSetup(t *testing.T) {
	start := time.Now()
	rootSetup(42, 1024)
	elapsed := time.Since(start)
	fmt.Println("timed: ", elapsed)
}

func TestCommitRoot(t *testing.T) {

	params := rootSetup(42, 10)

	scalVect := certToScalarVector(testCerts.certs[:10])
	commitment := rootCommit(params, scalVect)
	fmt.Println("commit: ", commitment)

}

func TestProveRoot(t *testing.T) {
	params := rootSetup(42, 10)
	scalVect := certToScalarVector(testCerts.certs[:10])
	//commitment := rootCommit(params, scalVect)
	proof := rootProveGen(params, scalVect, 4)
	fmt.Println("proof: ", proof)
}

func TestVerifyRoot(t *testing.T) {
	start := time.Now()
	fanout := 512
	params := rootSetup(42, fanout)
	elapsed := time.Since(start)
	fmt.Println("Time setup:", elapsed)
	start = time.Now()
	scalVect := certToScalarVector(testCerts.certs[:fanout])
	elapsed = time.Since(start)
	fmt.Println("certToScalVect time: ", elapsed)

	start = time.Now()
	commitment := rootCommit(params, scalVect)
	elapsed = time.Since(start)
	fmt.Println("time commit: ", elapsed)
	start = time.Now()
	proof := rootProveGen(params, scalVect, 4)
	elapsed = time.Since(start)
	fmt.Println("Time gen proof", elapsed)
	start = time.Now()
	didItWork := rootVerify(params, commitment, proof, scalVect[4], 4)
	elapsed = time.Since(start)
	fmt.Println("VerifyProof time", elapsed)
	fmt.Println("IT WORKKEDD: ", didItWork)

}
