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
		//{27},
	}
	scalVect := certToScalarVector(points)
	fmt.Println(scalVect)
	thePoly := realVectorToPoly(scalVect)
	fmt.Println("the Coefs!!!:", thePoly.coefficients)
	var k e.Scalar
	k.SetUint64(9)
	var x e.Scalar
	x.SetUint64(1)
	fmt.Println("THIS IS X ANAD SHOULD BE 1", x)
	answer := calcPoly(1, thePoly)
	fmt.Println("THE ANSWER: ", answer.IsEqual(&k))
	fmt.Println("this is 9 as Scalar from Uint", answer)

	//Test basic scalar math
	//var testVal1 e.Scalar
	//var testVal2 e.Scalar
	//testVal1.SetString("5")
	//testVal2.SetString("2")
	//fmt.Println(testVal1, testVal2)
	////testVal1.Neg()
	//testVal2.Inv(&testVal2)
	//fmt.Println(testVal1, testVal2)
	//testVal1.Mul(&testVal1, &testVal2)
	//fmt.Println(testVal1)
	//hej, err := strconv.ParseInt(answer.String(), 16, 64)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("i try the String method", hej)

	//var wat [][]byte
	//vectToPoly(wat)
}

func TestBasicScalarMath(t *testing.T) {
	//Test basic scalar math
	fmt.Println("Heeeeer under")
	var testVal1 e.Scalar
	var testVal2 e.Scalar
	testVal1.SetString("5")
	testVal2.SetString("2")
	fmt.Println(testVal1, testVal2)
	//testVal1.Neg()
	testVal2.Inv(&testVal2)
	fmt.Println(testVal1, testVal2)
	testVal1.Mul(&testVal1, &testVal2)
	fmt.Println(testVal1)
	fmt.Println("2.5*4=?")
	testVal2.SetString("4")
	testVal1.Mul(&testVal1, &testVal2)
	fmt.Println(testVal1)

	fmt.Println("Heeeeer over")
}
func TestSomethingNew(t *testing.T) {
	var fem e.Scalar
	var atten e.Scalar
	var minusOtte e.Scalar
	var x e.Scalar
	var answer e.Scalar
	fem.SetString("5")
	atten.SetString("26217937587563095239723870254092982918845276250263818911301829349969290592246")
	minusOtte.SetString("52435875175126190479447740508185965837690552500527637822603658699938581184505")
	//minusOtte.Neg()
	x.SetString("1")
	fmt.Println("minusOtte: ", minusOtte)
	answer.Add(&answer, &fem)
	atten.Mul(&atten, &x)
	answer.Add(&atten, &answer)
	x.Mul(&x, &x)
	minusOtte.Mul(&minusOtte, &x)
	answer.Add(&answer, &minusOtte)
	fmt.Println("f(1)=", answer)

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
	fmt.Println("verify Poly Returns:", verifyPoly(pk,commit, polynomial))

	index, fxo, witness := createWitness(pk, polynomial, uint64(1))

	fmt.Println("Very eval returns: ", verifyWitness(pk, commit, index, fxo, witness))

	open()
	fmt.Println("Testen løb igennen")

}
