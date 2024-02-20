package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestVerifyTree(t *testing.T) {
	fmt.Println("TestVerifyTree -  starting")
	certArray := loadCertificates("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result := verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	fmt.Println("TestVerifyCert -  starting")

	certArray := loadCertificates("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result := verifyNode(certArray[5], *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestTreeBuilder(t *testing.T) {
	fmt.Println("TestTreeBuilder -  starting")

	max := 1000
	min := 10
	for i := 0; i < 100; i++ {
		randNumb := rand.Intn(max-min) + min
		certArray := loadCertificates("testCerts/", randNumb)
		merkTree := BuildTree(certArray, 2)
		nodeToTest := rand.Intn(randNumb)
		result := verifyNode(certArray[nodeToTest], *merkTree)

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}

		result1 := verifyTree(certArray, *merkTree)
		if result1 != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}
