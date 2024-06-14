package verkletree

import (
	"fmt"
	"testing"

	combin "gonum.org/v1/gonum/stat/combin"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

func TestCreatedPolyEvalsCorrectly(t *testing.T) {
	fmt.Println("TestCreatedPolyEvalsCorrectly Running")
	points := [][]byte{
		{5},
		{15},
		{9},
		//{27},
	}
	scalVect := certToScalarVector(points)
	//fmt.Println(scalVect)
	var degreeComb [][][]int
	for k := len(points) - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(len(points), k-1))
	}
	divisorList := divisorCalculator(len(points), degreeComb)
	lagrangeBasisList := lagrangeBasisCalc(3, degreeComb, divisorList)
	thePoly := realVectorToPoly(scalVect, lagrangeBasisList)
	//fmt.Println("the Coefs!!!:", thePoly.coefficients)
	var k e.Scalar
	var x e.Scalar
	for i, p := range points {
		x.SetUint64(uint64(i))
		k.SetBytes(p)
		answer := calcPoly(uint64(i), thePoly)
		if answer.IsEqual(&k) == 0 {
			fmt.Println("Was incorrect for ", i, "should have been", p, "but was", answer)
			t.Error("The poly evals incorrectly")
		}
	}
}
func TestQuotientPoly(t *testing.T) {
	fmt.Println("TestQuotientPoly Running")
	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	scalVect := certToScalarVector(points)
	//fmt.Println(scalVect)
	var testScalar e.Scalar
	var degreeComb [][][]int
	for k := len(points) - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(len(points), k-1))
	}
	divisorList := divisorCalculator(len(points), degreeComb)
	lagrangeBasisList := lagrangeBasisCalc(4, degreeComb, divisorList)
	thePoly := realVectorToPoly(scalVect, lagrangeBasisList)
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
		t.Error("it is very wrong")
	}
	//fmt.Println("Succes")

}

func TestCommit(t *testing.T) {
	fmt.Println("TestCommit Running")
	points := [][]byte{
		{5},
		{15},
		{9},
		//{27},
	}
	pk := setup(3)

	var degreeComb [][][]int
	for k := len(points) - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(len(points), k-1))
	}
	divisorList := divisorCalculator(len(points), degreeComb)
	lagrangeBasisList := lagrangeBasisCalc(3, degreeComb, divisorList)

	polynomial := certVectorToPolynomial(points, lagrangeBasisList)
	commit := commit(pk, polynomial)
	verifyBool := verifyPoly(pk, commit, polynomial)
	if !verifyBool {
		t.Error("verify Poly was very incorrect")
	}

	witness := createWitness(pk, polynomial, uint64(1))
	verifyBool = verifyWitness(pk, commit, witness)
	if !verifyBool {
		t.Error("verify witness was very incorrect")
	}
}
