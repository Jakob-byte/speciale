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
	certArray := loadCertificates("AllCertsOneFile20000", 2)
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
	for i := 26; i <= 26; i++ {
		fmt.Println("Current fanout: ", i)
		testAmount := 5
		start := time.Now()
		fanOut := i
		pk := setup(4, fanOut)
		certArray := loadCertificates("AllCertsOneFile20000", 2)
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
	certArray := loadCertificates("AllCertsOneFile20000", 2)

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
	certArray := loadCertificates("AllCertsOneFile20000", 1)
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

func TestTreeBuild2(t *testing.T) {
	fmt.Println("TestTreeBuild2 -  starting")
	fanout:= 15
	pk := setup(10, fanout)
	certArray := loadCertificates("AllCertsOneFile20000", 5)
	BuildTree(certArray, fanout, pk)
}

func TestMembershipProofTimes(t *testing.T) {
	fmt.Println("TestMemberShipProofTimes Running")
	testAmount := 5
	start := time.Now()
	fanOut := 20
	pk := setup(4, fanOut)
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	elapsed1 := time.Since(start)

	fmt.Println("time elapsed for loading certs, and setup : ", elapsed1, "seconds")

	start = time.Now()
	verkTree := BuildTree(certArray, fanOut, pk)
	elapsed2 := time.Since(start).Seconds() / float64(testAmount)
	fmt.Println("Built tree time : ", elapsed2, "seconds")

	var success bool
	indexToTime := 56
	certToWitness := certArray[indexToTime]

	start = time.Now()
	membershipProof := createMembershipProof(certToWitness, *verkTree)
	success = verifyMembershipProof(membershipProof, verkTree.pk)
	elapsed3 := time.Since(start).Milliseconds()
	if success != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
	}

	start = time.Now()
	membershipProof = createMembershipProof(certToWitness, *verkTree)
	success = verifyMembershipProof(membershipProof, verkTree.pk)
	elapsed4 := time.Since(start).Milliseconds()
	if success != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
	}

	start = time.Now()
	membershipProof = createMembershipProof(certToWitness, *verkTree)
	success = verifyMembershipProof(membershipProof, verkTree.pk)
	elapsed5 := time.Since(start).Milliseconds()
	if success != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
	}

	if elapsed3 < elapsed4 {
		t.Errorf("Result was incorrect, got: ") //%t, want: %t.", elapsed3, elapsed4)
	}

	fmt.Println("VerifyTree time First time: ", elapsed3, "ms")
	fmt.Println("VerifyTree time Second time: ", elapsed4, "ms")
	fmt.Println("VerifyTree time third time: ", elapsed5, "ms")

}

func TestNewReadCertFunc(t *testing.T) {
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	//fmt.Println(certArray)
	if false {
		fmt.Println(certArray)
	}
}

// Benchmark/party time!!!!!!!
var testCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFIle20000", 5),
}

var table = []struct {
	fanOut int
	tree   verkleTree
}{
	//{input: 1}, Doesn't work for some reasone :D
	//{fanOut: 2, tree: *BuildTree(testCerts.certs, 2, setup(10, 2))},
	//{fanOut: 3, tree: *BuildTree(testCerts.certs, 3, setup(10, 3))},
	//{fanOut: 4, tree: *BuildTree(testCerts.certs, 4, setup(10, 4))},
	//{fanOut: 5, tree: *BuildTree(testCerts.certs, 5, setup(10, 5))},
	//{fanOut: 6, tree: *BuildTree(testCerts.certs, 6, setup(10, 6))},
	//{fanOut: 7, tree: *BuildTree(testCerts.certs, 7, setup(10, 7))},
	//{fanOut: 8, tree: *BuildTree(testCerts.certs, 8, setup(10, 8))},
	//{fanOut: 9, tree: *BuildTree(testCerts.certs, 9, setup(10, 9))},
	//{fanOut: 10, tree: *BuildTree(testCerts.certs, 10, setup(10, 10))},
	//{fanOut: 11, tree: *BuildTree(testCerts.certs, 11, setup(10, 11))},
	//{fanOut: 12, tree: *BuildTree(testCerts.certs, 12, setup(10, 12))},
	//{fanOut: 13, tree: *BuildTree(testCerts.certs, 13, setup(10, 13))},
	//{fanOut: 14, tree: *BuildTree(testCerts.certs, 14, setup(10, 14))},
	//{fanOut: 15, tree: *BuildTree(testCerts.certs, 15, setup(10, 15))},
	//{fanOut: 16, tree: *BuildTree(testCerts.certs, 16,setup(10,16))},
	//{fanOut: 17, tree: *BuildTree(testCerts.certs, 17,setup(10,17))},
	//{fanOut: 18, tree: *BuildTree(testCerts.certs, 18,setup(10,18))},
	//{fanOut: 19, tree: *BuildTree(testCerts.certs, 19,setup(10,19))},
	//{fanOut: 20, tree: *BuildTree(testCerts.certs, 20,setup(10,20))},
	//{fanOut: 25, tree: *BuildTree(testCerts.certs, 25,setup(10,25))},
}

func BenchmarkBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkBuildTreeTime Running")
	b.ResetTimer()
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			fanOut := v.fanOut
			pk := setup(4, fanOut)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BuildTree(testCerts.certs, fanOut, pk)

				//result := verifyTree(certArray, *verkTree, pk)

				//if result != true {
				//	b.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
				//}
			}
		})
	}
}

func BenchmarkVerifyNode(b *testing.B) {
	fmt.Println("BenchmarkVerifyNode Running")
	b.ResetTimer()

	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				verifyNode(testCerts.certs[i], v.tree)
			}
		})
	}
}
