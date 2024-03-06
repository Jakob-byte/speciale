package verkletree

import (
	"fmt"
	"testing"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

func TestCreatedPolyEvalsCorrectly(t *testing.T) {
	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	scalVect := certToScalarVector(points)
	//fmt.Println(scalVect)
	thePoly := realVectorToPoly(scalVect)
	//fmt.Println("the Coefs!!!:", thePoly.coefficients)
	var k e.Scalar
	var x e.Scalar
	for i, p := range points {
		x.SetUint64(uint64(i))
		k.SetBytes(p)
		answer := calcPoly(uint64(i), thePoly)
		if answer.IsEqual(&k) == 0 {
			fmt.Println("Was incorrect for ", i, "should have been", p, "but was", answer)
			panic("The poly evals incorrectly")
		}

	}
}
func TestQuotientPoly(t *testing.T) {
	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	scalVect := certToScalarVector(points)
	//fmt.Println(scalVect)
	var testScalar e.Scalar
	thePoly := realVectorToPoly(scalVect)
	quotientPoly := quotientOfPoly(thePoly, 2)
	var invertThing e.Scalar
	invertThing.SetString("3")
	invertThing.Inv(&invertThing)
	testScalar.SetString("20")
	testScalar.Mul(&testScalar, &invertThing)
	//testScalar.Neg()
	//fmt.Println("THIS IS -6.666", testScalar)
	//
	//testScalar.SetString("8")
	//testScalar.Neg()
	//fmt.Println("THIS IS -6.66", testScalar)
	//
	if false {
		fmt.Println(quotientPoly)
		panic("it is very wrong")
	}
	//fmt.Println("Succes")

}

func TestCommit(t *testing.T) {
	points := [][]byte{
		{5},
		{15},
		{9},
		//{27},
	}
	pk := setup(3, 3)

	polynomial := certVectorToPolynomial(points)
	commit := commit(pk, polynomial)
	fmt.Println("verify Poly Returns:", verifyPoly(pk, commit, polynomial))

	index, fxo, witness := createWitness(pk, polynomial, uint64(1))

	fmt.Println("Very eval returns: ", verifyWitness(pk, commit, index, fxo, witness))
	open()
	fmt.Println("Testen løb igennen")

}
