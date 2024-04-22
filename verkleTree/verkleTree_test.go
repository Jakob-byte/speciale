package verkletree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var testCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFIle20000", 100000),
}

var table = []struct {
	fanOut int
	tree   verkleTree
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
	{fanOut: 10, tree: *BuildTree(testCerts.certs, 10, setup(10, 10), 8)},
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

func TestBuildTreeAndVerifyTree(t *testing.T) {
	fmt.Println("TestBuildTreeAndVerifyTree Running ")
	points := [][]byte{
		{5},
		{15},
		{19},
		{27},
	}
	fanOut := 2
	pk := setup(1, fanOut)
	verk := BuildTree(points, fanOut, pk, 1)

	didItVerify := verifyTree(points, *verk, pk, 500)
	if !didItVerify {
		panic("Did not verify tree as expected")
	}
}

func TestVerifyNode(t *testing.T) {
	fmt.Println("TestVerifyNode Running")

	points := [][]byte{
		{5},
		{15},
		{19},
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

func TestMembershipProof2(t *testing.T) {
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
	pk := setup(1, fanOut)
	verk := BuildTree(points, fanOut, pk, 2)
	mp := createMembershipProof(points[2], *verk)
	didPointVerify := verifyMembershipProof(mp, pk)
	//fmt.Println("memberShipProof", mp)
	//fmt.Println("leafs", verk.leafs[0])
	if !didPointVerify {
		panic("point did not verify as expected")
	}
}

func TestMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestMembershipProofRealCerts Running")
	max := len(testCerts.certs)
	fanOut := 10
	pk := setup(10, fanOut)
	verkTree := BuildTree(testCerts.certs, fanOut, pk, 500)

	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max)
		mp := createMembershipProof(testCerts.certs[randNumb], *verkTree)
		didPointVerify := verifyMembershipProof(mp, pk)
		if didPointVerify != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
			break
		}
	}
}

func TestNegativeMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestNegativeMembershipProofRealCerts Running")
	fanOut := 10
	certToTest := testCerts.certs[30242]
	pk1 := setup(10, fanOut)
	pk2 := setup(10, fanOut)
	verkTree1 := BuildTree(testCerts.certs[:50000], fanOut, pk1, 500)
	verkTree2 := BuildTree(testCerts.certs[:50000], fanOut, pk2, 500)
	memProof := createMembershipProof(certToTest, *verkTree1)
	if verifyMembershipProof(memProof, verkTree2.pk) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestNegativeVerifyTree(t *testing.T) {
	fmt.Println("TestNegativeVerifyTree Running")
	fanOut := 10
	pk1 := setup(10, fanOut)
	pk2 := setup(10, fanOut)
	verkTree1 := BuildTree(testCerts.certs, fanOut, pk1, 8)
	if verifyTree(testCerts.certs, *verkTree1, pk2, 8) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestDifferentAmountOfThreadsDoesNotMakeDifferentTrees(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreadsDoesNotMakeDifferentTrees -  starting")
	fanOut := 10
	pk1 := setup(10, fanOut)
	verkTree1 := BuildTree(testCerts.certs, fanOut, pk1, 8)
	verkTree2 := BuildTree(testCerts.certs, fanOut, pk1, 800)

	if !verkTree1.Root.ownVectorCommit.IsEqual(&verkTree2.Root.ownVectorCommit) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

// Testing for the best amount of threads for the pc building the tree.
// 8 for Ryzen 7 4700u
// 4 for intel core i7 8th gen
// Probably always equal to the amount of cores on the CPU.
func TestDifferentAmountOfThreads(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreads -  starting")
	fanOut := 20
	pk := setup(42, fanOut)

	for threads := 2; threads < 20; threads++ {
		start := time.Now()
		BuildTree(testCerts.certs, fanOut, pk, threads)
		elapsed := time.Since(start)
		fmt.Println("Time elapsed making tree with fanout: ", fanOut, " and threads:", threads, "is: ", elapsed)
	}
}

// Good for testing and bugfixin' new code.
func TestRealCertificatesTime(t *testing.T) {
	fmt.Println("TestRealCertificatesTime Running")
	for i := 14; i <= 14; i++ {
		fmt.Println("Current fanout: ", i)
		testAmount := 5
		start := time.Now()
		fanOut := i
		pk := setup(4, fanOut)
		start = time.Now()
		var verkTree *verkleTree
		for i := 0; i < testAmount; i++ {
			verkTree = BuildTree(testCerts.certs, fanOut, pk, 500)
		}
		elapsed2 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("Built tree time : ", elapsed2, "seconds")

		start = time.Now()
		var result bool
		for i := 0; i < testAmount; i++ {
			result = verifyTree(testCerts.certs, *verkTree, pk, 500)
		}
		elapsed3 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("VerifyTree time : ", elapsed3, "seconds")

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}

//This test no longer works, as we now sort the certificates.
//func TestDumbUpdateLeafButEvil(t *testing.T) {
//	fmt.Println("TestDumbUpdateLeafButEvil Running")
//	fanOut := 10
//	pk := setup(4, fanOut)
//	certArray := testCerts.certs
//
//	verkTree := BuildTree(certArray, fanOut, pk, 500)
//
//	oneCert := loadOneCert("baguetteCert.crt")
//
//	newVerkTree, succes := dumbUpdateLeaf(*verkTree, oneCert, certArray[10])
//
//	if !succes {
//		t.Error("dumbUpdate func failed failed.")
//	}
//	certArray[10] = oneCert
//	itWorked := verifyTree(certArray, newVerkTree, newVerkTree.pk, 500)
//	if !itWorked {
//		t.Error("Failed verifying dumb-updated tree")
//	}
//}

// Tests whether the JSON converter works correctly, by comparing the membership proofs from before and after using it.
func TestJsonConverter(t *testing.T) {
	fmt.Println("TestJsonConverter Running")
	fanOut := 25
	pk := setup(30, fanOut)
	verkTree := BuildTree(testCerts.certs, fanOut, pk, 500)

	mp := createMembershipProof(testCerts.certs[1], *verkTree)
	bytesss := mp.Commitments[1].Bytes()
	mp.Commitments[1].SetBytes(bytesss)

	didPointVerify := verifyMembershipProof(mp, pk)
	if didPointVerify != true {
		t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
	}
	jsonTest := createJsonOfMembershipProof(mp)

	retrievedMP := retrieveMembershipProofFromJson(jsonTest)

	didPointVerify = verifyMembershipProof(retrievedMP, pk)
	if didPointVerify != true {
		t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
	}
	//fmt.Println("WHat is going on", retrievedMP)
	for i := range mp.Witnesses {
		if !mp.Witnesses[i].W.IsEqual(&retrievedMP.Witnesses[i].W) {
			fmt.Println("They are not equal, send help!")
		}
	}
}

//Insert doesn't work as we have them sorted now.
//func TestInsertSimple(t *testing.T) {
//	fmt.Println("TestInsertSimple Running")
//	fanOut := 3
//	pk := setup(4, fanOut)
//	certArray := loadCertificates("AllCertsOneFile20000", 1)
//	verkTree := BuildTree(certArray, fanOut, pk, 500)
//	baguetteCert := loadOneCert("baguetteCert.crt")
//	newTree, itWorked := insertLeaf(baguetteCert, *verkTree)
//	if !itWorked {
//		t.Error("inserting baguette certificate into tree failed ")
//	}
//	certArray = append(certArray, baguetteCert)
//	verifiedTree := verifyTree(certArray, newTree, pk, 500)
//
//	if !verifiedTree {
//		t.Error("Somehow insertLeaf worked, but it was not added to the tree. At least not correctly. Have a nice day.")
//	}
//}

// TODO is a bad test.
//
//	func TestMembershipProofTimes(t *testing.T) {
//		fmt.Println("TestMemberShipProofTimes Running - bad test")
//		start := time.Now()
//		fanOut := 15
//		pk := setup(4, fanOut)
//		elapsed1 := time.Since(start)
//
//		fmt.Println("time elapsed for loading certs, and setup : ", elapsed1)
//
//		start = time.Now()
//		verkTree := BuildTree(testCerts.certs, fanOut, pk, 500)
//		elapsed2 := time.Since(start)
//		fmt.Println("Built tree time : ", elapsed2)
//
//		var success bool
//		indexToTime := 5342
//		certToWitness := testCerts.certs[indexToTime]
//
//		start = time.Now()
//		membershipProof := createMembershipProof(certToWitness, *verkTree)
//		success = verifyMembershipProof(membershipProof, verkTree.pk)
//		elapsed3 := time.Since(start).Milliseconds()
//		if success != true {
//			t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
//		}
//
//		start = time.Now()
//		membershipProof = createMembershipProof(certToWitness, *verkTree)
//		success = verifyMembershipProof(membershipProof, verkTree.pk)
//		elapsed4 := time.Since(start).Milliseconds()
//		if success != true {
//			t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
//		}
//
//		start = time.Now()
//		membershipProof = createMembershipProof(certToWitness, *verkTree)
//		success = verifyMembershipProof(membershipProof, verkTree.pk)
//		elapsed5 := time.Since(start).Milliseconds()
//		if success != true {
//			t.Errorf("Result was incorrect, got: %t, want: %t.", success, true)
//		}
//
//		if elapsed3 < elapsed4 {
//			t.Errorf("Result was incorrect, got: ") //%t, want: %t.", elapsed3, elapsed4)
//		}
//		//fmt.Println("MembereshipProof", membershipProof)
//		fmt.Println("VerifyTree time First time: ", elapsed3, "ms")
//		fmt.Println("VerifyTree time Second time: ", elapsed4, "ms")
//		fmt.Println("VerifyTree time third time: ", elapsed5, "ms")
//
// }

// Function to test the size of the witness/proofs needed for different fanouts of the tree.
func TestSizeOfWitnesses(t *testing.T) {
	fmt.Println("TestSizeOfWitnesses Running")

	randInt := rand.Intn(len(testCerts.certs))
	randomCert := testCerts.certs[randInt]
	witnessList := make([][]byte, len(table))
	for i, v := range table {
		witnessList[i] = createJsonOfMembershipProof(createMembershipProof(randomCert, v.tree))
	}

	for i, v := range witnessList {
		fmt.Println("At fanout ", table[i].fanOut, " and ", len(testCerts.certs), " certificates, the size of the witness is", len(v))
	}

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
				BuildTree(testCerts.certs, fanOut, pk, 500)
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

// go test -bench=BenchmarkCreateMembershipProof -run=^a -benchtime=1000x -benchmem  -timeout 99999s | tee verkVerifyMemProofBench.txt
func BenchmarkCreateMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkVerifyNode Running")

	randomCerts := make([][]byte, 10000)

	for k := range randomCerts {
		randInt := rand.Intn(len(testCerts.certs))
		randomCerts[k] = testCerts.certs[randInt]
	}

	b.ResetTimer()
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				createMembershipProof(randomCerts[i], v.tree)
			}
		})
	}
}

// To run this test
// go test -bench=BenchmarkVerifyMembershipProof -run=^a -benchtime=1000x -benchmem  -timeout 99999s | tee merkVerifyWitnessBench.txt
func BenchmarkVerifyMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkVerifyMembershipProof Running")
	testAmount := 10000 //Change if you change -benchtime=10000x
	certsToTest := make([][]byte, testAmount)
	witnesses := make([][]membershipProof, len(table))
	start := time.Now()
	for o := range witnesses {
		witnesses[o] = make([]membershipProof, testAmount)
	}
	elapsed := time.Since(start)
	fmt.Println("Time spent after witness 1", elapsed)

	//Get certs to test
	for k := range testAmount {
		randInt := rand.Intn(len(testCerts.certs))
		certsToTest[k] = testCerts.certs[randInt]
		//(0, len(testCerts.certs))
	}
	elapsed = time.Since(start)
	fmt.Println("Time spent after witness 2", elapsed)
	//get proofs from the different trees
	for i, v := range table {
		for k := range testAmount {
			witnesses[i][k] = createMembershipProof(certsToTest[k], v.tree)
		}
	}
	elapsed = time.Since(start)
	fmt.Println("Time spent after witness 3", elapsed)

	b.ResetTimer()
	for o, v := range table {
		b.Run(fmt.Sprintf("fanOut: %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				verifyMembershipProof(witnesses[o][i], v.tree.pk)
			}
		})
	}
}
