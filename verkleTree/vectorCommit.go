package verkletree

import (
	"fmt"
)

func setup() int {
	fmt.Println("42 is the answer")
	return 42
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

func useEverything() {
	verifyEval()
	createWitness()
	verifyPoly()
	open()
	commit()
	setup()
	useEverything()
}
