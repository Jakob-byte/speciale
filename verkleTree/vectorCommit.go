package verkletree

import (
	"crypto/rand"
	"fmt"
	"math/big"

	//	"math/big"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

type PK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
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
		fmt.Println(gList[i].String())
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
func commit(pk PK, vectToCommit [][]byte) e.G1 {
	var commitment e.G1
	polyCoefs := vectToPolySike(vectToCommit) //TODO make a poly before calling commit.
	//	polyCoefs := vectToPoly(input)
	phiScalar := new(e.Scalar)
	var cToBe e.G1
	for i, phi := range polyCoefs {
		phiScalar.SetBytes(phi.Bytes())
		cToBe.ScalarMult(phiScalar, &pk.alphaG1s[i])
		commitment.Add(&cToBe, &commitment) //TODO Should there be a "mult" here somehow, as that is what is described in the original KZG paper.
	}

	return commitment
}

func open() int { //TODO fiks den aka. lav den
	return 0
}

func verifyPoly(pk PK, commitmentToVerify e.G1, vectToCommit [][]byte) bool {
	commitment := commit(pk, vectToCommit)
	return commitment.IsEqual(&commitmentToVerify)
}

func createWitness(pk PK, vectToCommit [][]byte, index int) (int, []byte, e.G1) {

	phiI := vectToCommit[index] //TODO to call to polynomial, when we get that to work
	w := commit(pk, vectToCommit)
	return index, phiI, w
}

func verifyEval(pk PK, commitment e.G1, index int, vectToCommit [][]byte, witness e.G1) bool {
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
	fxi.SetBytes(vectToCommit[index])
	rSide2.Exp(rSide2, fxi)
	rSide1.Mul(rSide1, rSide2)
	return lSide.IsEqual(rSide1)
}

func vectToPolySike(input [][]byte) []big.Int {
	var result []big.Int
	for _, v := range input {
		var intBig big.Int
		result = append(result, *intBig.SetBytes(v))
	}
	return result
}

//func vectToPoly(input [][]byte) func(int) big.Int {
//	n := len(input)
//	fmt.Println("Yay:", n)
//	lagrange := func(x int) big.Int {
//		var result big.Int
//		for j := 0; j < n; j++ {
//			lagrangeRes := 1
//			for m := 0; m < n; m++ {
//				if m != j {
//					lagrangeRes *= (x - m) / (j - m)
//				}
//			}
//			bigstuff := big.NewInt(0)
//			bigstuff.SetBytes(input[j])
//			bigstuff2 := big.NewInt(int64(lagrangeRes))
//			result = *result.Add(&result, bigstuff.Mul(bigstuff, bigstuff2))
//		}
//
//		return result
//	}
//	return lagrange
//}
