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
	rootsetup(42, 1024)
	elapsed := time.Since(start)
	fmt.Println("timed: ", elapsed)
}

func TestCommitRoot(t *testing.T) {

	params := rootsetup(42, 10)

	scalVect := certToScalarVector(testCerts.certs[:10])
	commitment := rootCommit(params, scalVect)
	fmt.Println("commit: ", commitment)

}

func TestProveRoot(t *testing.T) {
	params := rootsetup(42, 10)
	scalVect := certToScalarVector(testCerts.certs[:10])
	//commitment := rootCommit(params, scalVect)
	proof := rootProveGen(params, scalVect, 4)
	fmt.Println("proof: ", proof)
}

func TestVerifyRoot(t *testing.T) {

	params := rootsetup(42, 16)

	scalVect := certToScalarVector(testCerts.certs[:16])
	fmt.Println(len(testCerts.certs[:16]))
	commitment := rootCommit(params, scalVect)
	proof := rootProveGen(params, scalVect, 4)
	didItWork := rootVerify(params, commitment, proof, scalVect[4], 4)
	fmt.Println("IT WORKKEDD!!!!!!!!!!!!!!!!!: ", didItWork)

}
