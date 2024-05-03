package verkletree

import (
	"fmt"
	"testing"
	"time"
)

// TODO rewrite these tests?
func TestSetup(t *testing.T) {
	start := time.Now()
	optimizedSetup(42, 1024)
	elapsed := time.Since(start)
	fmt.Println("timed: ", elapsed)
}

func TestCommitRoot(t *testing.T) {

	params := optimizedSetup(42, 10)

	scalVect := certToScalarVector(testCerts.certs[:10])
	commitment := optimizedCommit(params, scalVect)
	fmt.Println("commit: ", commitment)

}

func TestProveRoot(t *testing.T) {
	params := optimizedSetup(42, 10)
	scalVect := certToScalarVector(testCerts.certs[:10])
	//commitment := rootCommit(params, scalVect)
	proof := optimizedProveGen(params, scalVect, 4)
	fmt.Println("proof: ", proof)
}

func TestVerifyRoot(t *testing.T) {
	start := time.Now()
	fanout := 512
	params := optimizedSetup(42, fanout)
	elapsed := time.Since(start)
	fmt.Println("Time setup:", elapsed)
	start = time.Now()
	scalVect := certToScalarVector(testCerts.certs[:fanout])
	elapsed = time.Since(start)
	fmt.Println("certToScalVect time: ", elapsed)

	start = time.Now()
	commitment := optimizedCommit(params, scalVect)
	elapsed = time.Since(start)
	fmt.Println("time commit: ", elapsed)
	start = time.Now()
	proof := optimizedProveGen(params, scalVect, 4)
	elapsed = time.Since(start)
	fmt.Println("Time gen proof", elapsed)
	start = time.Now()
	didItWork := optimizedVerify(params, commitment, proof, scalVect[4], 4)
	elapsed = time.Since(start)
	fmt.Println("VerifyProof time", elapsed)
	fmt.Println("IT WORKKEDD: ", didItWork)

}
