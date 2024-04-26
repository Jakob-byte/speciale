package verkletree

import (
	"fmt"
	"testing"
	"math/rand"
	"time"
)

var rootTestCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFIle20000", 1000000),
}

var rootTable = []struct {
	fanOut int
	tree   rootVerkleTree
}{
	//{input: 1}, Doesn't work for some reasone :D
	//{fanOut: 2, tree: *BuildTree(testCerts.certs, 2, setup(10, 2), 8)},
	//{fanOut: 3, tree: *BuildTree(testCerts.certs, 3, setup(10, 3), 8)},
	//{fanOut: 4, tree: *BuildTree(testCerts.certs, 4, setup(10, 4), 8)},
	//{fanOut: 5, tree: *BuildTree(testCerts.certs, 5, setup(10, 5), 8)},
	//{fanOut: 6, tree: *BuildTree(testCerts.certs, 6, setup(10, 6), 8)},
	//{fanOut: 7, tree: *BuildTree(testCerts.certs, 7, setup(10, 7), 8)},
	//{fanOut: 8, tree: *BuildTree(testCerts.certs, 8, setup(10, 8), 8)},
	//{fanOut: 9, tree: *BuildTree(testCerts.certs, 9, setup(10, 9), 8)},
	{fanOut: 10, tree: *rootBuildTree(testCerts.certs, 10, rootSetup(10, 10), 8)},
	//{fanOut: 11, tree: *BuildTree(testCerts.certs, 11, setup(10, 11), 8)},
	//{fanOut: 12, tree: *BuildTree(testCerts.certs, 12, setup(10, 12), 8)},
	//{fanOut: 13, tree: *BuildTree(testCerts.certs, 13, setup(10, 13), 8)},
	//{fanOut: 14, tree: *BuildTree(testCerts.certs, 14, setup(10, 14), 8)},
	//{fanOut: 15, tree: *BuildTree(testCerts.certs, 15, setup(10, 15), 8)},
	//{fanOut: 16, tree: *BuildTree(testCerts.certs, 16, setup(10, 16), 8)},
	//{fanOut: 17, tree: *BuildTree(testCerts.certs, 17, setup(10, 17), 8)},
	//{fanOut: 18, tree: *BuildTree(testCerts.certs, 18, setup(10, 18), 8)},
	//{fanOut: 19, tree: *BuildTree(testCerts.certs, 19, setup(10, 19), 8)},
	//{fanOut: 20, tree: *BuildTree(testCerts.certs, 20, setup(10, 20), 8)},
	//{fanOut: 25, tree: *BuildTree(testCerts.certs, 25, setup(10, 25), 8)},
}

var fanOuts = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

func TestRootBuildTreeAndVerifyTree(t *testing.T) {
	fmt.Println("TestBuildTreeAndVerifyTree Running ")
	points := [][]byte{
		{5},
		{15},
		{19},
		{27},
	}
	fanOut := 2
	pk := rootSetup(1, fanOut)
	verk := rootBuildTree(points, fanOut, pk, 1)

	didItVerify := rootVerifyTree(points, *verk, pk, 8)
	if !didItVerify {
		panic("Did not verify tree as expected")
	}
}

func TestRootMembershipProof(t *testing.T) {
	fmt.Println("verifyMemberShip Running")

	points := [][]byte{
		{5},
		{15},
		{19},
		{27},
		{30},
		{40},
		{50},
		{60},
	}
	fanOut := 2
	pk := rootSetup(1, fanOut)
	verk := rootBuildTree(points, fanOut, pk, 2)
	mp := rootCreateMembershipProof(points[2], *verk)
	didPointVerify := rootVerifyMembershipProof(mp, pk)
	//fmt.Println("memberShipProof", mp)
	//fmt.Println("leafs", verk.leafs[0])
	if !didPointVerify {
		panic("point did not verify as expected")
	}
}
func TestRootMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestMembershipProofRealCerts Running")
	max := len(testCerts.certs)
	fanOut := 10
	pk := rootSetup(10, fanOut)
	verkTree := rootBuildTree(testCerts.certs, fanOut, pk, 8)

	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max)
		mp := rootCreateMembershipProof(testCerts.certs[randNumb], *verkTree)
		didPointVerify := rootVerifyMembershipProof(mp, pk)
		if didPointVerify != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
			break
		}
	}
}

func TestRootNegativeMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestNegativeMembershipProofRealCerts Running")
	fanOut := 10
	certToTest := testCerts.certs[30242]
	pk1 := rootSetup(10, fanOut)
	pk2 := rootSetup(10, fanOut)
	verkTree1 := rootBuildTree(testCerts.certs[:50000], fanOut, pk1, 8)
	verkTree2 := rootBuildTree(testCerts.certs[:50000], fanOut, pk2, 8)
	memProof := rootCreateMembershipProof(certToTest, *verkTree1)
	if rootVerifyMembershipProof(memProof, verkTree2.pk) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestRootNegativeVerifyTree(t *testing.T) {
	fmt.Println("TestNegativeVerifyTree Running")
	fanOut := 10
	pk1 := rootSetup(10, fanOut)
	pk2 := rootSetup(10, fanOut)
	verkTree1 := rootBuildTree(testCerts.certs, fanOut, pk1, 8)
	if rootVerifyTree(testCerts.certs, *verkTree1, pk2, 8) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestRootDifferentAmountOfThreadsDoesNotMakeDifferentTrees(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreadsDoesNotMakeDifferentTrees -  starting")
	fanOut := 10
	pk1 := rootSetup(10, fanOut)
	verkTree1 := rootBuildTree(testCerts.certs, fanOut, pk1, 8)
	verkTree2 := rootBuildTree(testCerts.certs, fanOut, pk1, 800)

	if !verkTree1.Root.ownVectorCommit.IsEqual(&verkTree2.Root.ownVectorCommit) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}
func TestRootDifferentAmountOfThreads(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreads -  starting")
	fanOut := 10
	pk := rootSetup(42, fanOut)

	for threads := 1; threads < 20; threads++ {
		start := time.Now()
		rootBuildTree(testCerts.certs, fanOut, pk, threads)
		elapsed := time.Since(start)
		fmt.Println("Time elapsed making tree with fanout: ", fanOut, " and threads:", threads, "is: ", elapsed)
	}
}

func TestRootRealCertificatesTime(t *testing.T) {
	fmt.Println("TestRealCertificatesTime Running")
	for i := 14; i <= 14; i++ {
		fmt.Println("Current fanout: ", i)
		testAmount := 5
		start := time.Now()
		fanOut := i
		pk := rootSetup(4, fanOut)
		start = time.Now()
		var verkTree *rootVerkleTree
		for i := 0; i < testAmount; i++ {
			verkTree = rootBuildTree(testCerts.certs, fanOut, pk, 8)
		}
		elapsed2 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("Built tree time : ", elapsed2, "seconds")

		start = time.Now()
		var result bool
		for i := 0; i < testAmount; i++ {
			result = rootVerifyTree(testCerts.certs, *verkTree, pk, 8)
		}
		elapsed3 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("VerifyTree time : ", elapsed3, "seconds")

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}

// go test -bench=BenchmarkRootBuildTreeTime -run=^a -benchtime=1x -benchmem  -timeout 99999s | tee verkrootBuildTreeBench.txt
func BenchmarkRootBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkRootBuildTreeTime Running")
	b.ResetTimer()
	for _, v := range fanOuts.v {
		b.Run(fmt.Sprintf("fanOut: %d", v), func(b *testing.B) {
			fanOut := v
			pk := rootSetup(4, fanOut)
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootBuildTree(testCerts.certs, fanOut, pk, 12)
			}
		})
	}
}