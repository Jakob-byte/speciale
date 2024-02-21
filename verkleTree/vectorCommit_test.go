package verkletree

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestSomethingIGuess2(t *testing.T) {
	fmt.Println(rand.Intn(45))
	verifyEval()
	createWitness()
	verifyPoly()
	open()
	commit()
	setup(10, 10)
	var wat [][]byte
	vectToPoly(wat)
}
