package verkletree

import (
	"fmt"
	"testing"
)

func TestSetupComitVerifyPoly(t *testing.T) {
	var pk PK
	//var inp [][]byte
	points := []int{
		2,
		5,
		11,
	}
	pk = setup(10, 10)
	commit := commit(pk, points)
	//verifyPoly(pk, commit, inp)
	fmt.Println(commit)
	fmt.Println(verifyPoly(pk, commit, points))
	//verifyEval(pk, commit, 1, inp, commit)
	//createWitness(pk, inp, 1)
	//
	//open()
	//var wat [][]byte
	//vectToPoly(wat)
}
func TestWitness(t *testing.T) {
	var pk PK
	//var inp [][]byte
	points := []int{
		2,
		5,
		11,
	}
	pk = setup(10, 10)
	commit := commit(pk, points)
	fmt.Println(commit)
	lagrange := vectToPoly(points)
	index, _, witness := createWitness(pk, points, *lagrange, 1)
	fmt.Println(verifyEval(pk, commit, index, points, witness))
}
