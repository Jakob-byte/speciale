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
	result := verifyNode(certArray[5], *merkTree, 2)
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
		result := verifyNode(certArray[nodeToTest], *merkTree, 2)

		if result != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", result, true)
			break
		}

		result1 := verifyTree(certArray, *merkTree)
		if result1 != true {
			t.Errorf("Result from VerifyTree was incorrect, got: %t, want: %t.", result1, true)
			break
		}
	}
}

func TestDifferentFanOuts(t *testing.T) {
	fmt.Println("TestDifferentFanOuts -  starting")
	max := 500
	min := 100
	maxFan := 100
	minFan := 2
	for i := 0; i < 100; i++ {
		randNumb := rand.Intn(max-min) + min
		fanNumb := rand.Intn(maxFan-minFan) + minFan
		certArray := loadCertificates("testCerts/", randNumb)
		merkTree := BuildTree(certArray, fanNumb)
		nodeToTest := rand.Intn(randNumb)
		result := verifyNode(certArray[nodeToTest], *merkTree, fanNumb)

		if result != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", result, true)
			break
		}

		result1 := verifyTree(certArray, *merkTree)
		if result1 != true {
			t.Errorf("Result from VerifyTree was incorrect, got: %t, want: %t.", result1, true)
			break
		}
	}
}
