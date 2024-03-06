package verkletree

import (
	"fmt"
	"testing"
	"time"
)

func TestBuildTreeAndVerifyTree(t *testing.T) {
	fmt.Println("TestBuildTreeAndVerifyTree Running ")
	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	fanOut := 2
	pk := setup(1, fanOut)
	verk := BuildTree(points, fanOut, pk)

	didItVerify := verifyTree(points, *verk, pk)
	if !didItVerify {
		panic("Did not verify tree as expected")
	}
}

func TestVerifyNode(t *testing.T) {
	fmt.Println("TestVerifyNode Running")

	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	fanOut := 2
	pk := setup(1, fanOut)
	verk := BuildTree(points, fanOut, pk)
	didNodeVerify := verifyNode(points[2], *verk, pk)
	
	if !didNodeVerify{
		panic("Node did not verify as expected")
	} 
}

func TestRealCertificates(t *testing.T) {
	start := time.Now()
	fmt.Println("TestRealCertificates Running")
	fanOut := 10
	pk := setup(4, fanOut)
	certArray := loadCertificates("testCerts/")
	elapsed1 := time.Since(start)
	fmt.Println("Elapsed1 : ", elapsed1)
	
	start = time.Now()
	verkTree := BuildTree(certArray, fanOut, pk)
	elapsed2 := time.Since(start)
	fmt.Println("Elapsed2 : ", elapsed2)
	
	start = time.Now()
	result := verifyTree(certArray, *verkTree, pk)
	elapsed3 := time.Since(start)
	fmt.Println("Elapsed3 : ", elapsed3)


	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}