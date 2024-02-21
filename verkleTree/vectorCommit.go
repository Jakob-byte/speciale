package verkletree

import (
	"crypto/rand"
	"fmt"

	//"math"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

type PK struct {
	g1s []e.G1
	g2s []e.G2
}

// what bit security do we have or bls12381
// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
func setup(security int, t int) PK {
	fmt.Println("42 is the answer")
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	gList := make([]e.G1, t+1)
	at := new(e.Scalar)
	at.Set(a)
	gList[0] = *g1
	for i := 1; i <= t; i++ {
		gList[i].ScalarMult(at, g1)
		fmt.Println(gList[i].String())
		at.Mul(at, a)
	}

	ag2 := new(e.G2)
	ag2.ScalarMult(a, g2)
	pk := PK{
		g1s: gList,
		g2s: []e.G2{*g2, *ag2},
	}
	return pk
}

func commit() int {

	return 0
}

func open() int {
	return 0
}

func verifyPoly() int {
	return 0
}

func createWitness() int {
	return 0
}

func verifyEval() int {
	return 0
}

func vectToPoly(input [][]byte) func(int) int {
	n := len(input)
	fmt.Println("Yay:", n)
	lagrange := func(x int) int {
		result := 0
		for j := 0; j < n; j++ {
			lagrangeRes := 1
			for m := 0; m < n; m++ {
				if m != j {
					lagrangeRes *= (x - m) / (j - m)
				}
			}
			result += input[j] * lagrangeRes
		}

		return result
	}
	return lagrange
}
