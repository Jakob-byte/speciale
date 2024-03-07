package verkletree

import (
	"fmt"
	"testing"
	"time"
	"math/rand"
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

	if !didNodeVerify {
		panic("Node did not verify as expected")
	}
}

func TestMembershipProof(t *testing.T) {
	fmt.Println("verifyMemberShip Running")

	points := [][]byte{
		{5},
		{15},
		{9},
		{27},
	}
	fanOut := 2
	pk := setup(1, fanOut)
	verk := BuildTree(points, fanOut, pk)
	mp := createMembershipProof(points[2], *verk)
	didPointVerify := verifyMembershipProof(mp, pk)
	
	if !didPointVerify {
		panic("point did not verify as expected")
	}
}

func TestMembershipProofRealCerts(t *testing.T){
	fmt.Println("TestMembershipProofRealCerts Running")
	max := 300
	fanOut := 10
	pk := setup(10, fanOut)
	certArray := loadCertificates("testCerts/", max)
	verkTree := BuildTree(certArray, fanOut,pk)

	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max)
		mp := createMembershipProof(certArray[randNumb], *verkTree)
		didPointVerify := verifyMembershipProof(mp, pk)
		if didPointVerify != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
			break
		}
	}

}

func TestRealCertificatesTime(t *testing.T) {
	testAmount := 10
	start := time.Now()
	fmt.Println("TestRealCertificatesTime Running")
	fanOut := 10
	pk := setup(4, fanOut)
	certArray := loadCertificates("testCerts/")
	elapsed1 := time.Since(start)
	
	fmt.Println("time elapsed for loading certs, and setup : ", elapsed1)

	start = time.Now()
	var verkTree *verkleTree
	for i:= 0; i<testAmount; i++{
		verkTree = BuildTree(certArray, fanOut, pk)
	}
	elapsed2 := time.Since(start).Seconds() / float64(testAmount)
	fmt.Println("Built tree time : ", elapsed2)

	start = time.Now()
	var result bool
	for i:= 0; i<testAmount; i++{
		result = verifyTree(certArray, *verkTree, pk)
	}
	elapsed3 := time.Since(start).Seconds() / float64(testAmount)
	fmt.Println("VerifyTree time : ", elapsed3)

	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}
