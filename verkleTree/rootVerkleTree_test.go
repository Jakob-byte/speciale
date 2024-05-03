package verkletree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var numThreads = 13
var witnessBool = false
var rootTestCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFile20000", 100000),
}

var fanOuts = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

// TODO edit this to some good amount of certAmount Diego said this was fine
var certAmount = struct {
	c []int
}{
	c: []int{10000, 20000, 40000, 80000, 160000, 240000, 480000, 1000000},
}
var rootTable = []struct {
	fanOut int
	tree   rootVerkleTree
}{
	//takes roughly 7 minutes, don't worry
	//{input: 1}, Doesn't work for some reasone :D
	//{fanOut: 2, tree: *rootBuildTree(rootTestCerts.certs, 2, rootSetup(10, 2), witnessBool, numThreads)},
	//{fanOut: 4, tree: *rootBuildTree(rootTestCerts.certs, 4, rootSetup(10, 4), witnessBool, numThreads)},
	//{fanOut: 8, tree: *rootBuildTree(rootTestCerts.certs, 8, rootSetup(10, 8), witnessBool, numThreads)},
	//{fanOut: 16, tree: *rootBuildTree(rootTestCerts.certs, 16, rootSetup(10, 16), witnessBool, numThreads)},
	{fanOut: 32, tree: *rootBuildTree(rootTestCerts.certs, 32, rootSetup(10, 32), witnessBool, numThreads)},
	//{fanOut: 64, tree: *rootBuildTree(rootTestCerts.certs, 64, rootSetup(10, 64), witnessBool, numThreads)},
	//{fanOut: 128, tree: *rootBuildTree(rootTestCerts.certs, 128, rootSetup(10, 128), witnessBool, numThreads)},
	//{fanOut: 256, tree: *rootBuildTree(rootTestCerts.certs, 256, rootSetup(10, 256), witnessBool, numThreads)},
	//{fanOut: 512, tree: *rootBuildTree(rootTestCerts.certs, 512, rootSetup(10, 512), witnessBool, numThreads)},
	//{fanOut: 1024, tree: *rootBuildTree(rootTestCerts.certs, 1024, rootSetup(10, 1024), witnessBool, numThreads)},
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
	verk := rootBuildTree(points, fanOut, pk, witnessBool, numThreads)

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
	verk := rootBuildTree(points, fanOut, pk, witnessBool, numThreads)
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
	max := len(rootTestCerts.certs)
	fanOut := 10
	pk := rootSetup(10, fanOut)
	verkTree := rootBuildTree(rootTestCerts.certs, fanOut, pk, witnessBool, numThreads)

	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max)
		mp := rootCreateMembershipProof(rootTestCerts.certs[randNumb], *verkTree)
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
	certToTest := rootTestCerts.certs[30242]
	pk1 := rootSetup(10, fanOut)
	pk2 := rootSetup(10, fanOut)
	verkTree1 := rootBuildTree(rootTestCerts.certs[:50000], fanOut, pk1, witnessBool, numThreads)
	verkTree2 := rootBuildTree(rootTestCerts.certs[:50000], fanOut, pk2, witnessBool, numThreads)
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
	verkTree1 := rootBuildTree(rootTestCerts.certs, fanOut, pk1, witnessBool, numThreads)
	if rootVerifyTree(rootTestCerts.certs, *verkTree1, pk2, numThreads) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestRootDifferentAmountOfThreadsDoesNotMakeDifferentTrees(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreadsDoesNotMakeDifferentTrees -  starting")
	fanOut := 10
	pk1 := rootSetup(10, fanOut)
	verkTree1 := rootBuildTree(rootTestCerts.certs, fanOut, pk1, witnessBool, 8)
	verkTree2 := rootBuildTree(rootTestCerts.certs, fanOut, pk1, witnessBool, 800)

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
		rootBuildTree(rootTestCerts.certs, fanOut, pk, witnessBool, threads)
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
			verkTree = rootBuildTree(rootTestCerts.certs, fanOut, pk, witnessBool, numThreads)
		}
		elapsed2 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("Built tree time : ", elapsed2, "seconds")

		start = time.Now()
		var result bool
		for i := 0; i < testAmount; i++ {
			result = rootVerifyTree(rootTestCerts.certs, *verkTree, pk, 8)
		}
		elapsed3 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("VerifyTree time : ", elapsed3, "seconds")

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}

// Tests whether the JSON converter works correctly, by comparing the membership proofs from before and after using it.
func TestRootJsonConverter(t *testing.T) {
	fmt.Println("TestJsonConverter Running")
	fanOut := 2

	pk := rootSetup(30, fanOut)

	verkTree := rootBuildTree(rootTestCerts.certs, fanOut, pk, witnessBool, numThreads)

	mp := rootCreateMembershipProof(rootTestCerts.certs[1], *verkTree)
	//fmt.Println("Before set bytes:")
	//fmt.Println(mp.Commitments[1])
	//copuy := mp.Commitments[1]
	bytesss := mp.Commitments[1].Bytes()
	mp.Commitments[1].SetBytes(bytesss)
	//fmt.Println("After set bytes")
	//fmt.Println(mp.Commitments[1])
	//fmt.Println("But are they are equal: ", copuy.IsEqual(&mp.Commitments[1]))

	didPointVerify := rootVerifyMembershipProof(mp, pk)
	if didPointVerify != true {
		t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
	}
	jsonTest := rootCreateJsonOfMembershipProof(mp)

	retrievedMP := rootRetrieveMembershipProofFromJson(jsonTest)
	//fmt.Println("retrieved mp", retrievedMP)
	didPointVerify = rootVerifyMembershipProof(retrievedMP, pk)
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

// Function to test the size of the witness/proofs needed for different fanouts of the tree.
func TestRootSizeOfWitnesses(t *testing.T) {
	fmt.Println("TestSizeOfWitnesses Running")

	randInt := rand.Intn(len(rootTestCerts.certs))
	randomCert := rootTestCerts.certs[randInt]
	fmt.Println("Set the random cert")
	witnessList := make([][]byte, len(rootTable))
	fmt.Println("Made witness list")
	for i, v := range rootTable {
		witnessList[i] = rootCreateJsonOfMembershipProof(rootCreateMembershipProof(randomCert, v.tree))
	}
	fmt.Println("Finished making json proofs")
	for i, v := range witnessList {
		fmt.Println("At fanout ", rootTable[i].fanOut, " and ", len(rootTestCerts.certs), " certificates, the size of the witness is", len(v))
	}

}

// go test -bench=BenchmarkRootSetupPkTime -run=^a -benchtime=10x -benchmem  -timeout 99999s | tee verkRootSetupPkBench.txt

func BenchmarkRootSetupPkTime(b *testing.B) {
	fmt.Println("BenchmarkRootSetupPkTime Running")
	for _, v := range fanOuts.v {
		b.Run(fmt.Sprintf("fanOut: %d", v), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootSetup(4, v)
			}
		})
	}

}

// go test -bench=BenchmarkRootBuildTreeTime -run=^a -benchtime=1x -benchmem  -timeout 99999s | tee verkRootBuildTreeBench.txt
func BenchmarkRootBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkRootBuildTreeTime Running")
	b.ResetTimer()
	for _, v := range fanOuts.v {
		b.Run(fmt.Sprintf("fanOut: %d", v), func(b *testing.B) {
			pk := rootSetup(4, v)
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootBuildTree(rootTestCerts.certs, v, pk, witnessBool, numThreads)
			}
		})
	}
}

// go test -bench=BenchmarkRootVerifyNode -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee verkRootVerifyMemProofBench.txt
func BenchmarkRootVerifyNode(b *testing.B) {
	fmt.Println("BenchmarkRootVerifyNode Running")
	b.ResetTimer()
	testAmount := 10 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(rootTestCerts.certs))
		randomCerts[k] = rootTestCerts.certs[randInt]
	}

	for _, v := range rootTable {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootVerifyNode(randomCerts[i], v.tree)
			}
		})
	}
}

// go test -bench=BenchmarkRootCreateMembershipProof -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee verkRootVerifyMemProofBench.txt
func BenchmarkRootCreateMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkRootCreateMembershipProof Running")
	testAmount := 10 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(rootTestCerts.certs))
		randomCerts[k] = rootTestCerts.certs[randInt]
	}

	b.ResetTimer()
	for _, v := range rootTable {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootCreateMembershipProof(randomCerts[i], v.tree)
			}
		})
	}
}

// TODO NOT FINISHED JUST SAME AS ABOVE
// TODO how do we make these trees do we build them all in this case or earlier??
// build for fanouts for amountofCerts save in list then generate membershiproof and state what fanout/amount we are in?
// go test -bench=BenchmarkRootCreateMembershipProofVaryingAmountOfCerts -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee VerkRootCreateMembershipProofBench.txt
func BenchmarkRootCreateMembershipProofVaryingAmountOfCerts(b *testing.B) {
	fmt.Println("BenchmarkRootCreateMembershipProof Running")
	testAmount := 10 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(rootTestCerts.certs))
		randomCerts[k] = rootTestCerts.certs[randInt]
	}

	b.ResetTimer()
	for _, v := range rootTable {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootCreateMembershipProof(randomCerts[i], v.tree)
			}
		})
	}
}

// To run this test
// go test -bench=BenchmarkRootVerifyMembershipProof -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee VerkRootVerifyWitnessBench.txt
func BenchmarkRootVerifyMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkRootVerifyMembershipProof Running")
	testAmount := 10 //Change if you change -benchtime=10000x
	certsToTest := make([][]byte, testAmount)
	witnesses := make([][]rootMembershipProof, len(rootTable))
	start := time.Now()
	for o := range witnesses {
		witnesses[o] = make([]rootMembershipProof, testAmount)
	}
	elapsed := time.Since(start)
	fmt.Println("Time spent after witness 1", elapsed)

	//Get certs to test
	for k := range testAmount {
		randInt := rand.Intn(len(rootTestCerts.certs))
		certsToTest[k] = rootTestCerts.certs[randInt]
	}
	elapsed = time.Since(start)
	fmt.Println("Time spent after witness 2", elapsed)
	//get proofs from the different trees
	for i, v := range rootTable {
		for k := range testAmount {
			witnesses[i][k] = rootCreateMembershipProof(certsToTest[k], v.tree)
		}
	}
	elapsed = time.Since(start)
	fmt.Println("Time spent after witness 3", elapsed)

	b.ResetTimer()
	for o, v := range rootTable {
		b.Run(fmt.Sprintf("fanOut: %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rootVerifyMembershipProof(witnesses[o][i], v.tree.pk)
			}
		})
	}
}

// go test -bench=BenchmarkRootDifferentAmountOfCertsBuild -benchtime=1x -run=^a -benchmem  -timeout 99999s | tee BenchmarkRootDifferentAmountOfCertsBuild.txt
func BenchmarkRootDifferentAmountOfCertsBuild(b *testing.B) {
	fmt.Println("BenchmarkRootDifferentAmountOfCertsBuild -  starting")
	//fanOuts := []int{2} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
	for _, fan := range fanOuts.v {
		pk := rootSetup(4, fan)
		b.ResetTimer()
		for _, amountOfCerts := range certAmount.c {
			b.Run(fmt.Sprintf("fanOut: %d and amountOfCerts %d", fan, amountOfCerts), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					rootBuildTree(rootTestCerts.certs[:amountOfCerts], fan, pk, witnessBool, numThreads)
				}
			})
		}
	}
}

// go test -bench=BenchmarkRootDifferentAmountOfThreads -benchtime=3x -run=^a -benchmem  -timeout 99999s | tee BenchmarkRootDifferentAmountOfThreads.txt
func BenchmarkRootDifferentAmountOfThreads(b *testing.B) {
	fmt.Println("BenchmarkRootDifferentAmountOfThreads -  starting")
	fanOuts := []int{32} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}

	for _, fan := range fanOuts {
		pk := rootSetup(4, fan)
		b.ResetTimer()
		for threads := 8; threads < 20; threads++ {
			b.Run(fmt.Sprintf("fanOut: %d and threads %d", fan, threads), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					rootBuildTree(rootTestCerts.certs, fan, pk, witnessBool, threads)

				}
			})
		}
	}
}
