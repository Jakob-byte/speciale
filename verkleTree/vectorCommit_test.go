package verkletree

import (
	"testing"
)

func TestSomethingIGuess2(t *testing.T) {
	var pk PK
	var inp [][]byte
	commit := commit(pk, inp)
	verifyPoly(pk, commit, inp)

	verifyEval(pk, commit, 1, inp, commit)
	createWitness(pk, inp, 1)

	open()
	setup(10, 10)
	//var wat [][]byte
	//vectToPoly(wat)
}
