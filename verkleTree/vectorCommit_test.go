package verkletree

import (
	"fmt"
	"testing"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

func TestSomethingIGuess2(t *testing.T) {
	//var pk PK
	//var inp [][]byte
	//commit := commit(pk, inp)
	//verifyPoly(pk, commit, inp)
	//
	//verifyEval(pk, commit, 1, inp, commit)
	//createWitness(pk, inp, 1)
	//
	//open()
	//setup(10, 10)
	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	scalVect := certToScalarVector(points)
	fmt.Println(scalVect)
	thePoly := realVectorToPoly(scalVect)
	fmt.Print(thePoly.coefficients)
	var k e.Scalar
	k.SetUint64(9)
	var x e.Scalar
	x.SetUint64(2)
	answer := calcPoly(x, thePoly)
	fmt.Println("THE ANSWER: ", answer.IsEqual(&k))
	fmt.Println("this is 9 as Scalar from Uint")
	//var wat [][]byte
	//vectToPoly(wat)
}
