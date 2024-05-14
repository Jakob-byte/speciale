package verkletree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var numThreads = 8
var witnessBool = false
var optimizedTestCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFile20000", 10000000), // TODO increase ALOT! :)
}

var fanOuts = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

var certAmount = struct {
	c []int
}{
	c: []int{1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000, 10000000}, //TODO change back
}
var optimizedTable = []struct {
	fanOut int
	tree   optimizedVerkleTree
}{
	//takes roughly 7 minutes, don't worry
	//{fanOut: 2, tree: *optimizedBuildTree(optimizedTestCerts.certs, 2, optimizedSetup(10, 2), witnessBool, numThreads)},
	//{fanOut: 4, tree: *optimizedBuildTree(optimizedTestCerts.certs, 4, optimizedSetup(10, 4), witnessBool, numThreads)},
	//{fanOut: 8, tree: *optimizedBuildTree(optimizedTestCerts.certs, 8, optimizedSetup(10, 8), witnessBool, numThreads)},
	{fanOut: 16, tree: *optimizedBuildTree(optimizedTestCerts.certs, 16, optimizedSetup(10, 16), witnessBool, numThreads)},
	//{fanOut: 32, tree: *optimizedBuildTree(optimizedTestCerts.certs, 32, optimizedSetup(10, 32), witnessBool, numThreads)},
	//{fanOut: 64, tree: *optimizedBuildTree(optimizedTestCerts.certs, 64, optimizedSetup(10, 64), witnessBool, numThreads)},
	//{fanOut: 128, tree: *optimizedBuildTree(optimizedTestCerts.certs, 128, optimizedSetup(10, 128), witnessBool, numThreads)},
	//{fanOut: 256, tree: *optimizedBuildTree(optimizedTestCerts.certs, 256, optimizedSetup(10, 256), witnessBool, numThreads)},
	//{fanOut: 512, tree: *optimizedBuildTree(optimizedTestCerts.certs, 512, optimizedSetup(10, 512), witnessBool, numThreads)},
	//{fanOut: 1024, tree: *optimizedBuildTree(optimizedTestCerts.certs, 1024, optimizedSetup(10, 1024), witnessBool, numThreads)},
}

func TestOptimizedBuildTreeAndVerifyTree(t *testing.T) {
	fmt.Println("TestBuildTreeAndVerifyTree Running ")
	points := [][]byte{
		{5},
		{15},
		{19},
		{27},
	}
	fanOut := 2
	pk := optimizedSetup(1, fanOut)
	verk := optimizedBuildTree(points, fanOut, pk, witnessBool, numThreads)

	didItVerify := optimizedVerifyTree(points, *verk, pk, 8)
	if !didItVerify {
		panic("Did not verify tree as expected")
	}
}

func TestOptimizedMembershipProof(t *testing.T) {
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
	pk := optimizedSetup(1, fanOut)
	verk := optimizedBuildTree(points, fanOut, pk, witnessBool, numThreads)
	mp := optimizedCreateMembershipProof(points[2], *verk)
	didPointVerify := optimizedVerifyMembershipProof(mp, pk)
	//fmt.Println("memberShipProof", mp)
	//fmt.Println("leafs", verk.leafs[0])
	if !didPointVerify {
		panic("point did not verify as expected")
	}
}
func TestOptimizedMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestMembershipProofRealCerts Running")
	max := len(optimizedTestCerts.certs)
	fanOut := 10
	pk := optimizedSetup(10, fanOut)
	verkTree := optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk, witnessBool, numThreads)

	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max)
		mp := optimizedCreateMembershipProof(optimizedTestCerts.certs[randNumb], *verkTree)
		didPointVerify := optimizedVerifyMembershipProof(mp, pk)
		if didPointVerify != true {
			t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
			break
		}
	}
}

func TestOptimizedNegativeMembershipProofRealCerts(t *testing.T) {
	fmt.Println("TestNegativeMembershipProofRealCerts Running")
	fanOut := 10
	certToTest := optimizedTestCerts.certs[30242]
	pk1 := optimizedSetup(10, fanOut)
	pk2 := optimizedSetup(10, fanOut)
	verkTree1 := optimizedBuildTree(optimizedTestCerts.certs[:50000], fanOut, pk1, witnessBool, numThreads)
	verkTree2 := optimizedBuildTree(optimizedTestCerts.certs[:50000], fanOut, pk2, witnessBool, numThreads)
	memProof := optimizedCreateMembershipProof(certToTest, *verkTree1)
	if optimizedVerifyMembershipProof(memProof, verkTree2.pk) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestOptimizedNegativeVerifyTree(t *testing.T) {
	fmt.Println("TestNegativeVerifyTree Running")
	fanOut := 10
	pk1 := optimizedSetup(10, fanOut)
	pk2 := optimizedSetup(10, fanOut)
	verkTree1 := optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk1, witnessBool, numThreads)
	if optimizedVerifyTree(optimizedTestCerts.certs, *verkTree1, pk2, numThreads) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

func TestOptimizedDifferentTreesAmountOfThreadsDoesNotMakeDifferentTrees(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreadsDoesNotMakeDifferentTrees -  starting")
	fanOut := 10
	pk1 := optimizedSetup(10, fanOut)
	verkTree1 := optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk1, witnessBool, 8)
	verkTree2 := optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk1, witnessBool, 800)

	if !verkTree1.Root.ownVectorCommit.IsEqual(&verkTree2.Root.ownVectorCommit) {
		t.Error("Accepted the memebershipproof, even though the pk was wrong. Send assitance!")
	}
}

// This test is currently modified to see how long it takes to gen tress with/without proofs
// Previously used to test how many threads were optimal for the CPU
func TestOptimizedDifferentAmountOfCertsWithWithoutProofs(t *testing.T) {
	fmt.Println("TestOptimizedDifferentAmountOfCertsWithWithoutProofs -  starting")
	fanOut := 2
	pk := optimizedSetup(42, fanOut)

	for amountOfCerts := 100000; amountOfCerts <= 1000000; amountOfCerts += 100000 {
		start := time.Now()
		optimizedBuildTree(optimizedTestCerts.certs[:amountOfCerts], fanOut, pk, true, 8)
		elapsed := time.Since(start)
		fmt.Println("Time elapsed making tree with fanout: ", fanOut, " and threads:", amountOfCerts, "is: ", elapsed, " and time now is: ", time.Now())
	}
}

func TestOptimizedRealCertificatesTime(t *testing.T) {
	fmt.Println("TestRealCertificatesTime Running")
	for i := 14; i <= 14; i++ {
		fmt.Println("Current fanout: ", i)
		testAmount := 5
		start := time.Now()
		fanOut := i
		pk := optimizedSetup(4, fanOut)
		start = time.Now()
		var verkTree *optimizedVerkleTree
		for i := 0; i < testAmount; i++ {
			verkTree = optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk, witnessBool, numThreads)
		}
		elapsed2 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("Built tree time : ", elapsed2, "seconds")

		start = time.Now()
		var result bool
		for i := 0; i < testAmount; i++ {
			result = optimizedVerifyTree(optimizedTestCerts.certs, *verkTree, pk, 8)
		}
		elapsed3 := time.Since(start).Seconds() / float64(testAmount)
		fmt.Println("VerifyTree time : ", elapsed3, "seconds")

		if result != true {
			t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		}
	}
}

// Tests whether the JSON converter works correctly, by comparing the membership proofs from before and after using it.
func TestOptimizedJsonConverter(t *testing.T) {
	fmt.Println("TestJsonConverter Running")
	fanOut := 2

	pk := optimizedSetup(30, fanOut)

	verkTree := optimizedBuildTree(optimizedTestCerts.certs, fanOut, pk, witnessBool, numThreads)

	mp := optimizedCreateMembershipProof(optimizedTestCerts.certs[1], *verkTree)
	//fmt.Println("Before set bytes:")
	//fmt.Println(mp.Commitments[1])
	//copuy := mp.Commitments[1]
	bytesss := mp.Commitments[1].Bytes()
	mp.Commitments[1].SetBytes(bytesss)
	//fmt.Println("After set bytes")
	//fmt.Println(mp.Commitments[1])
	//fmt.Println("But are they are equal: ", copuy.IsEqual(&mp.Commitments[1]))

	didPointVerify := optimizedVerifyMembershipProof(mp, pk)
	if didPointVerify != true {
		t.Errorf("Result from VerifyNode was incorrect, got: %t, want: %t.", didPointVerify, true)
	}
	jsonTest := optimizedCreateJsonOfMembershipProof(mp)

	retrievedMP := optimizedRetrieveMembershipProofFromJson(jsonTest)
	//fmt.Println("retrieved mp", retrievedMP)
	didPointVerify = optimizedVerifyMembershipProof(retrievedMP, pk)
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
// go test -run TestOptimizedSizeOfWitnesses | tee proofSizes.txt
func TestOptimizedSizeOfWitnesses(t *testing.T) {
	fmt.Println("TestOptimizedSizeOfWitnesses - starting")

	for _, w := range certAmount.c {
		randInt := rand.Intn(w)
		randomcertificate := optimizedTestCerts.certs[randInt]
		for _, v := range fanOuts.v {
			setttup := optimizedSetup(10, v)
			testTree := optimizedBuildTree(optimizedTestCerts.certs[:w], v, setttup, false, numThreads)
			size := optimizedCreateJsonOfMembershipProof(optimizedCreateMembershipProof(randomcertificate, *testTree))
			fmt.Println("fan-out: ", v, ", certificates: ", w, ". Witness/membershipProof size in bytes: ", len(size))
		}
	}
}

// go test -bench=BenchmarkRootSetupPkTime -run=^a -benchtime=10x -benchmem  -timeout 99999s | tee verkRootSetupPkBench.txt
func BenchmarkOptimizedSetupPkTime(b *testing.B) {
	fmt.Println("BenchmarkRootSetupPkTime Running")
	for _, v := range fanOuts.v {
		b.Run(fmt.Sprintf("fan-out: %d", v), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				optimizedSetup(4, v)
			}
		})
	}

}

// go test -bench=BenchmarkRootVerifyNode -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee verkRootVerifyMemProofBench.txt
func BenchmarkOptimizedVerifyNode(b *testing.B) {
	fmt.Println("BenchmarkRootVerifyNode Running")
	b.ResetTimer()
	testAmount := 10 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(optimizedTestCerts.certs))
		randomCerts[k] = optimizedTestCerts.certs[randInt]
	}
	b.ResetTimer()
	for _, v := range optimizedTable {

		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {

			for i := 0; i < b.N; i++ {
				optimizedVerifyNode(randomCerts[i], v.tree)
			}
		})
	}
}

// TODO NOT FINISHED JUST SAME AS ABOVE
// TODO how do we make these trees do we build them all in this case or earlier??
// build for fanouts for amountofCerts save in list then generate membershiproof and state what fanout/amount we are in?
// go test -bench=BenchmarkOptimizedvCreateMembershipProofVaryingAmountOfCerts -run=^a -benchtime=5x -benchmem  -timeout 99999s | tee VerkRootCreateMembershipProofBench.txt
func BenchmarkOptimizedvCreateMembershipProofVaryingAmountOfCerts(b *testing.B) {
	fmt.Println("BenchmarkRootCreateMembershipProof Running")
	testAmount := 10 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(optimizedTestCerts.certs))
		randomCerts[k] = optimizedTestCerts.certs[randInt]
	}

	b.ResetTimer()
	for _, v := range optimizedTable {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				optimizedCreateMembershipProof(randomCerts[i], v.tree)
			}
		})
	}
}

// go test -bench=BenchmarkOptimizedDifferentAmountOfCertsBuild -benchtime=1x -run=^a -benchmem  -timeout 99999s | tee BenchmarkOptimizedDifferentAmountOfCertsBuild.txt
func BenchmarkOptimizedDifferentAmountOfCertsBuild(b *testing.B) {
	fmt.Println("BenchmarkRootDifferentAmountOfCertsBuild -  starting")
	//fanOuts := []int{2} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
	for _, fan := range fanOuts.v {
		pk := optimizedSetup(4, fan)
		b.ResetTimer()
		for _, amountOfCerts := range certAmount.c {
			b.Run(fmt.Sprintf("fanOut: %d and amountOfCerts %d", fan, amountOfCerts), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					optimizedBuildTree(optimizedTestCerts.certs[:amountOfCerts], fan, pk, witnessBool, numThreads)
				}
			})
		}
	}
}

// go test -bench=BenchmarkOptimizedDifferentAmountOfThreads -benchtime=3x -run=^a -benchmem  -timeout 99999s | tee BenchmarkOptimizedDifferentAmountOfThreads.txt
func BenchmarkOptimizedDifferentAmountOfThreads(b *testing.B) {
	fmt.Println("BenchmarkRootDifferentAmountOfThreads -  starting")
	fanOuts := []int{32} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}

	for _, fan := range fanOuts {
		pk := optimizedSetup(4, fan)
		b.ResetTimer()
		for threads := 8; threads < 20; threads++ {
			b.Run(fmt.Sprintf("fanOut: %d and threads %d", fan, threads), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					optimizedBuildTree(optimizedTestCerts.certs, fan, pk, witnessBool, threads)

				}
			})
		}
	}
}

// TODO Run benchmark on server
// go test -bench=BenchmarkOptimizedBuildTreeTime -run=^a -benchtime=100x -benchmem  -timeout 9999999s | tee BenchmarkOptimizedBuildTreeTime.txt
func BenchmarkOptimizedBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkOptimizedBuildTreeTime - Starting")
	b.ResetTimer()
	for _, certs := range certAmount.c {
		for _, v := range fanOuts.v {
			b.ResetTimer()
			b.Run(fmt.Sprintf("fan-out: %d, Certs: %d", v, certs), func(b *testing.B) {
				pk := optimizedSetup(4, v)
				for i := 0; i < b.N; i++ {
					optimizedBuildTree(optimizedTestCerts.certs[:certs], v, pk, witnessBool, numThreads)
				}
			})
		}
	}
}

// TODO Run on server, and divide result from this test with 1000 to get actual average
// go test -bench=BenchmarkOptimizedCreateMembershipProof -run=^a -benchtime=200x -benchmem  -timeout 9999999s | tee BenchmarkOptimizedCreateMembershipProof.txt
func BenchmarkOptimizedCreateMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkOptimizedCreateMembershipProof - starting")
	testAmount := 200 //Change if you change -benchtime=10000x
	amountToAverageOver := 1000
	randomCerts := make([][]byte, testAmount)

	b.ResetTimer()
	for _, certs := range certAmount.c {
		for _, f := range fanOuts.v {
			pubParams := optimizedSetup(10, f)
			benchTree := optimizedBuildTree(optimizedTestCerts.certs[:certs], f, pubParams, false, numThreads)
			b.ResetTimer()
			b.Run(fmt.Sprintf("fan-out: %d, Certs: %d", f, certs), func(b *testing.B) {
				for range amountToAverageOver {
					b.StopTimer() //Stop timer, and generate 200 new certs to create proof for, then starts timer again.
					for k := range randomCerts {
						randInt := rand.Intn(len(optimizedTestCerts.certs[:certs]))
						randomCerts[k] = optimizedTestCerts.certs[randInt]
					}
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						optimizedCreateMembershipProof(randomCerts[i], *benchTree)
					}
				}
			})
		}
	}
}

// TODO Run on server
// go test -bench=BenchmarkOptimizedVerifyMembershipProof -run=^a -benchtime=1000x -benchmem  -timeout 9999999s | tee BenchmarkOptimizedVerifyMembershipProof.txt
func BenchmarkOptimizedVerifyMembershipProof(b *testing.B) {
	fmt.Println("BenchmarkOptimizedVerifyMembershipProof - starting")
	start := time.Now()
	testAmount := 1000 //Change if you change -benchtime=10000x
	certsToTest := make([][]byte, testAmount)
	witnesses := make([]optimizedMembershipProof, testAmount)

	elapsed := time.Since(start)
	fmt.Println("Time spent after witness 1", elapsed)

	elapsed = time.Since(start)
	fmt.Println("Time spent after witness 2", elapsed)
	//get proofs from the different trees
	for _, certs := range certAmount.c {
		for _, v := range fanOuts.v {
			params := optimizedSetup(10, v)
			benchTree := optimizedBuildTree(optimizedTestCerts.certs[:certs], v, params, false, numThreads)
			b.ResetTimer()
			b.Run(fmt.Sprintf("fan-out: %d, certs: %d", v, certs), func(b *testing.B) {
				// TODO stops timer and gets new random certs for this test. Should we do this?
				b.StopTimer()
				for k := range testAmount {
					randInt := rand.Intn(certs) // rand.Intn(len(optimizedTestCerts.certs[:certs]))
					certsToTest[k] = optimizedTestCerts.certs[randInt]
					witnesses[k] = optimizedCreateMembershipProof(certsToTest[k], *benchTree)
				}
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					optimizedVerifyMembershipProof(witnesses[i], params)
				}

			})
		}
	}
}

// TODO run on server
// This benchmark measures how the Create Membership proof time decreases after repeated queries.
// go test -bench=BenchmarkOptimizedCreateMemProofOverTime -run=^a -benchtime=1x -benchmem  -timeout 9999999s | tee BenchmarkOptimizedCreateMemProofOverTime.txt
func BenchmarkOptimizedCreateMemProofOverTime(b *testing.B) {
	fmt.Println("BenchmarkOptimizedCreateMemProofOverTime - starting")
	benchtime := 1                  //Should be same as benchtime
	testAmount := 10000 * benchtime //Change if you change -benchtime=10000x
	averageTimes := 100

	randomCerts := make([][]byte, testAmount)

	for _, certs := range certAmount.c {
		for _, f := range fanOuts.v {
			pk := optimizedSetup(10, f)
			benchTree := optimizedBuildTree(optimizedTestCerts.certs[:certs], f, pk, false, numThreads)
			for o := range averageTimes {
				for k := range testAmount {
					randInt := rand.Intn(certs)
					randomCerts[k] = optimizedTestCerts.certs[randInt]
				}
				for j := range testAmount {
					b.ResetTimer()
					b.Run(fmt.Sprintf("fan-out: %d, certs: %d, iteration: %d", f, certs, o), func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							optimizedCreateMembershipProof(randomCerts[j], *benchTree)
						}
					})
				}
			}
		}
	}
}
