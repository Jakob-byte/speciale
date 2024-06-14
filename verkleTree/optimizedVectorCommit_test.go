package verkletree

import (
	"testing"
)

// TODO rewrite these tests?

func TestOptimizedAllInVectorCommit(t *testing.T) {
	fanout := 512
	params := optimizedSetup(fanout)
	scalVect := certToScalarVector(testCerts.certs[:fanout])

	commitment := optimizedCommit(params, scalVect)
	proof := optimizedProofGen(params, scalVect, 4)

	didItWork := optimizedVerify(params, commitment, proof, scalVect[4], 4)

	if !didItWork {
		t.Error("There is a problem in the Optimized Vector Commit functions!")
	}
}
