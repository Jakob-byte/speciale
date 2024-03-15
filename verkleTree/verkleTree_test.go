package verkletree

import (
	"fmt"
	"math/rand"
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
	membershipProof := createMembershipProof(points[2], *verk)
	didNodeVerify := verifyMembershipProof(membershipProof, pk)

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

func TestMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestMembershipProofRealCerts Running")
	max := 300
	fanOut := 10
	pk := setup(10, fanOut)
	certArray := loadCertificates("testCerts/", max)
	verkTree := BuildTree(certArray, fanOut, pk)

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
	fmt.Println("TestRealCertificatesTime Running")
	for i := 20; i <= 30; i++ {
		fmt.Println("Current fanout: ", i)
		testAmount := 5
		start := time.Now()
		fanOut := i
		pk := setup(4, fanOut)
		certArray := loadCertificates("testCerts/")
		elapsed1 := time.Since(start)

		fmt.Println("time elapsed for loading certs, and setup : ", elapsed1, "seconds")

		start = time.Now()
		var verkTree *verkleTree
		for i := 0; i < testAmount; i++ {
			verkTree = BuildTree(certArray, fanOut, pk)
		}
		elapsed2 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("Built tree time : ", elapsed2, "seconds")

		start = time.Now()
		var result bool
		for i := 0; i < testAmount; i++ {
			result = verifyTree(certArray, *verkTree, pk)
		}
		elapsed3 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("VerifyTree time : ", elapsed3, "seconds")

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}

func TestDumbUpdateLeafButEvil(t *testing.T) {
	fmt.Println("TestDumbUpdateLeafButEvil Running")
	fanOut := 10
	pk := setup(4, fanOut)
	certArray := loadCertificates("testCerts/")

	verkTree := BuildTree(certArray, fanOut, pk)

	oneCert := loadOneCert("baguetteCert.crt")

	newVerkTree, succes := dumbUpdateLeaf(*verkTree, oneCert, certArray[10])

	if !succes {
		t.Error("dumbUpdate func failed failed.")
	}
	certArray[10] = oneCert
	itWorked := verifyTree(certArray, newVerkTree, newVerkTree.pk)
	if !itWorked {
		t.Error("Failed verifying dumb-updated tree")
	}
}

func TestInsertSimple(t *testing.T) {
	fmt.Println("TestInsertSimple Running")
	fanOut := 2
	pk := setup(4, fanOut)
	certArray := loadCertificates("testCerts/", 999)
	verkTree := BuildTree(certArray, fanOut, pk)
	baguetteCert := loadOneCert("baguetteCert.crt")
	newTree, itWorked := insertLeaf(baguetteCert, *verkTree)
	if !itWorked {
		t.Error("insert baguetteCert failed tree")
	}

	certArray = append(certArray, baguetteCert)
	verifiedTree := verifyTree(certArray, newTree, pk)

	if !verifiedTree {
		t.Error("Somehow insertLeaf worked, but it was not added to the tree. At least not correctly. Have a nice day.")
	}
}
