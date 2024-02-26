package verkletree

import (
	"bytes"
	"crypto/rand"
	"fmt"

	//	"math/big"
	"encoding/binary"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

type PK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
}

type Point struct {
	X float64
	Y float64
}

// what bit security do we have or bls12381
// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
func setup(security int, t int) PK {
	fmt.Println("42 is the answer")
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	gList := make([]e.G1, t)
	at := new(e.Scalar)
	at.Set(a)
	for i := 0; i < t; i++ {
		gList[i].ScalarMult(at, g1)
		//fmt.Println(gList[i].String())
		at.Mul(at, a)
	}

	ag2 := new(e.G2)
	ag2.ScalarMult(a, g2)
	pk := PK{
		g1:       *g1,
		alphaG1s: gList,
		g2:       *g2,
		alphaG2:  *ag2,
	}
	return pk
}

// TODO CHANGE VECTTOCOMMIT TO POLYNOMIAL!!!!!
func commit(pk PK, vectToCommit []int) e.G1 {
	commitment := pk.g1

	polyCoefs := vectToCommit //TODO make a poly before calling commit.
	//	polyCoefs := vectToPoly(input)
	phiScalar := new(e.Scalar)
	var cToBe e.G1
	for i, phi := range polyCoefs {
		buff := new(bytes.Buffer)
		err := binary.Write(buff, binary.LittleEndian, uint16(phi))
		if err != nil {
			fmt.Println(err)
		}

		intByteArray := buff.Bytes()

		phiScalar.SetBytes(intByteArray)

		cToBe.ScalarMult(phiScalar, &pk.alphaG1s[i])
		commitment.Add(&cToBe, &commitment) //TODO Should there be a "mult" here somehow, as that is what is described in the original KZG paper.
	}
	return commitment
}

// func open() int { //TODO fiks den aka. lav den
//
//		return 0
//	}
func verifyPoly(pk PK, commitmentToVerify e.G1, vectToCommit []int) bool {
	commitment := commit(pk, vectToCommit)
	fmt.Println("first commit", commitment)
	fmt.Println("second commit", commitmentToVerify)

	return commitment.IsEqual(&commitmentToVerify)
}

func createWitness(pk PK, vectToCommit []int, polynomial LagrangePolynomial, index int) (int, float64, e.G1) {

	phiI := polynomial.evaluate(float64(index)) //TODO to call to polynomial, when we get that to work
	w := commit(pk, vectToCommit)
	return index, phiI, w
}

func verifyEval(pk PK, commitment e.G1, index int, vectToCommit []int, witness e.G1) bool {
	lSide := e.Pair(&commitment, &pk.g2)

	// e(w, alpha * g2 - x0 * g2) * e(g1, g2) ^f(x_i)
	var alphaG2minusX0G2 e.G2
	xi := new(e.Scalar)
	xi.SetUint64(uint64(index))
	var xig2 e.G2
	xig2.ScalarMult(xi, &pk.g2)
	alphaG2minusX0G2.Add(&pk.alphaG2, &xig2)
	rSide1 := e.Pair(&witness, &alphaG2minusX0G2)
	rSide2 := e.Pair(&pk.g1, &pk.g2)
	fxi := new(e.Scalar)

	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.LittleEndian, uint16(vectToCommit[index]))
	if err != nil {
		fmt.Println(err)
	}

	intByteArray := buff.Bytes()

	fxi.SetBytes(intByteArray)
	rSide2.Exp(rSide2, fxi)
	rSide1.Mul(rSide1, rSide2)
	return lSide.IsEqual(rSide1)
}

//	func vectToPolySike(input [][]byte) []big.Int {
//		var result []big.Int
//		for _, v := range input {
//			var intBig big.Int
//			result = append(result, *intBig.SetBytes(v))
//		}
//		return result
//	}
type LagrangePolynomial struct {
	coefficients []float64
	evaluate     func(float64) float64
}

// vectToPoly calculates the Lagrange polynomial and returns its coefficients and evaluation function
func vectToPoly(points []int) *LagrangePolynomial {
	n := len(points)
	coefficients := make([]float64, n)
	// Inner function for evaluating the polynomial
	evaluate := func(x float64) float64 {
		result := 0.0
		for j := 0; j < n; j++ {
			lagrangeRes := 1.0
			//product := points[j].Y
			for m := 0; m < n; m++ {
				if m != j {
					lagrangeRes *= (x - float64(m)) / (float64(j) - float64(m))
				}
			}
			fmt.Println("lagrangeRes:", lagrangeRes)
			coefficients[j] = float64(points[j]) * lagrangeRes // Calculate and store coefficients
			result += coefficients[j]                          // Use coefficients for evaluation

		}
		return result
	}

	return &LagrangePolynomial{coefficients: coefficients, evaluate: evaluate}
}
