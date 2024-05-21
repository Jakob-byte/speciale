package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	//"time"
)

var numThreads = 32
var testCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFile20000", 10000000, numThreads), //TODO change back to 10 million
}

var fanOuts = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

// TODO edit this to some good amount of certAmount
var certAmount = struct {
	c []int
}{
	c: []int{1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000, 10000000}, // TODO fix before server benchmarks
}

var threads = struct {
	v []int
}{
	v: []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

var table = []struct {
	fanOut     int
	tree       merkleTree
	testFanout []int
}{
	//{input: 1, tree: *BuildTree(testCerts.certs, 1)}}, Doesn't work for some reasone :D  //TODO undo so we can test for different fanouts
	//	{fanOut: 2, tree: *BuildTree(testCerts.certs, 2, numThreads)},
	//{fanOut: 4, tree: *BuildTree(testCerts.certs, 4,numThreads)},
	//{fanOut: 8, tree: *BuildTree(testCerts.certs, 8,numThreads)},
	//{fanOut: 16, tree: *BuildTree(testCerts.certs, 16,numThreads)},
	//{fanOut: 32, tree: *BuildTree(testCerts.certs, 32,numThreads)},
	//{fanOut: 64, tree: *BuildTree(testCerts.certs, 64,numThreads)},
	//{fanOut: 128, tree: *BuildTree(testCerts.certs, 128,numThreads)},
	//{fanOut: 256, tree: *BuildTree(testCerts.certs, 256,numThreads)},
	//{fanOut: 512, tree: *BuildTree(testCerts.certs, 512,numThreads)},
	//{fanOut: 1024, tree: *BuildTree(testCerts.certs, 1024,numThreads)},
}

func TestLoadFunc(t *testing.T) {
	fmt.Println("TestLoadFunc -  starting")
	start := time.Now()
	certArray := loadCertificates("AllCertsOneFile20000", 20000, numThreads)
	elapsed := time.Since(start)
	fmt.Println("Time spent loading ", len(certArray), " certificates, it took: ", elapsed)
	if len(certArray) != 20000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 20000)
	}

	start = time.Now()
	certArray = loadCertificates("AllCertsOneFile20000", 40000, numThreads)
	elapsed = time.Since(start)
	fmt.Println("Time spent loading ", len(certArray), " certificates, it took: ", elapsed)
	if len(certArray) != 40000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 40000)
	}

	start = time.Now()
	certArray = loadCertificates("AllCertsOneFile20000", 60000, numThreads)
	elapsed = time.Since(start)
	fmt.Println("Time spent loading ", len(certArray), " certificates, it took: ", elapsed)
	if len(certArray) != 60000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 60000)
	}

	start = time.Now()
	certArray = loadCertificates("AllCertsOneFile20000", 80000, numThreads)
	elapsed = time.Since(start)
	fmt.Println("Time spent loading ", len(certArray), " certificates, it took: ", elapsed)
	if len(certArray) != 80000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 80000)
	}

	start = time.Now()
	certArray = loadCertificates("AllCertsOneFile20000", 1000000, numThreads)
	elapsed = time.Since(start)
	fmt.Println("Time spent loading ", len(certArray), " certificates, it took: ", elapsed)
	if len(certArray) != 1000000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 1000000)
	}
}

func TestVerifyTree(t *testing.T) {
	fmt.Println("TestVerifyTree -  starting")
	certArray := testCerts.certs
	merkTree := BuildTree(certArray, 2, numThreads)
	result := verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	fmt.Println("TestVerifyCert -  starting")

	certArray := testCerts.certs

	merkTree := BuildTree(certArray, 2, numThreads)
	//for i:= 0 ; i<10; i++ {
	//	fmt.Println(i, "hash at index , merkTree.leafs[i].parent.parent.ownHash)
	//}
	result := verifyNode(certArray[2], *merkTree)

	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		//fmt.Println(certArray[3204043959346])
	}

}

func TestNegativeVerifyCert(t *testing.T) {
	fmt.Println("TestNegativeVerifyCert -  starting")

	certArray := testCerts.certs[:10000]

	merkTree := BuildTree(certArray, 2, numThreads)
	//for i:= 0 ; i<10; i++ {
	//	fmt.Println(i, "hash at index , merkTree.leafs[i].parent.parent.ownHash)
	//}
	result := verifyNode(testCerts.certs[10001], *merkTree)

	if result {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, false)
		//fmt.Println(certArray[3204043959346])
	}
}

func TestNegativeWitnessTestWithDifferentCerts(t *testing.T) {
	fmt.Println("TestNegativeWitnessTestWithDifferentCerts Running")
	certsToTestOn1 := testCerts.certs[:10000]
	certsToTestOn2 := testCerts.certs[2000:12000]
	tree1 := BuildTree(certsToTestOn1, 10, numThreads)
	tree2 := BuildTree(certsToTestOn2, 10, numThreads)
	certToVerify := certsToTestOn1[2500]
	witness := createWitness(certToVerify, *tree1)
	if verifyWitness(certToVerify, witness, tree2.pubParams) {
		t.Error("Should not have output true, should be false.")
	}
}

func TestNegativeWitnessTestWithDifferentFanout(t *testing.T) {
	fmt.Println("TestNegativeWitnessTestWithDifferentFanout Running")
	certsToTestOn1 := testCerts.certs[:10000]
	certsToTestOn2 := testCerts.certs[:10000]
	tree1 := BuildTree(certsToTestOn1, 2, numThreads)
	tree2 := BuildTree(certsToTestOn2, 4, numThreads)
	certToVerify := certsToTestOn1[2500]
	witness := createWitness(certToVerify, *tree1)
	if verifyWitness(certToVerify, witness, tree2.pubParams) {
		t.Error("Should not have output true, should be false.")
	}
}

func TestTreeBuilder(t *testing.T) {
	fmt.Println("TestTreeBuilder -  starting")
	max := 100000
	min := 100
	fanMin := 2
	fanMax := 100
	threadMin := 1
	threadMax := 1000
	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max-min) + min
		randFan := rand.Intn(fanMax-fanMin) + fanMin
		randThread := rand.Intn(threadMax-threadMin) + threadMin
		certArray := loadCertificates("AllCertsOneFile20000", randNumb, numThreads)
		merkTree := BuildTree(certArray, randFan, randThread)
		nodeToTest := rand.Intn(len(certArray))
		result := verifyNode(certArray[nodeToTest], *merkTree)

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

func TestNegativeTreeVerify(t *testing.T) {
	fmt.Println("TestNegativeTreeVerify -  starting")
	certs := testCerts.certs[:100000]
	tree1 := BuildTree(certs, 2, numThreads)
	if verifyTree(testCerts.certs[20000:120000], *tree1) {
		t.Errorf("Verified the tree, although the certs were different")
	}
}

// Hej - goddav
func TestDifferentFanOuts(t *testing.T) {
	fmt.Println("TestDifferentFanOuts -  starting")
	max := 500
	min := 100
	maxFan := 100
	minFan := 2
	for i := 1; i < 2; i++ {
		randNumb := rand.Intn(max-min) + min
		fanNumb := rand.Intn(maxFan-minFan) + minFan
		certArray := testCerts.certs
		merkTree := BuildTree(certArray, fanNumb, numThreads)
		nodeToTest := rand.Intn(randNumb)
		result := verifyNode(certArray[nodeToTest], *merkTree)

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

func TestJsonConverterInTree(t *testing.T) {
	fmt.Println("TestJsonConverter Running")
	certToVerify := testCerts.certs[40123]
	witness := createWitness(certToVerify, table[0].tree)
	if !verifyWitness(certToVerify, witness, table[0].tree.pubParams) {
		t.Error("Should have been in the tree, what went wrong? (before conversion)")
	}

	jsooon := genJsonWitness(witness)
	witnessAgain := getWitnessFromJson(jsooon)
	if !verifyWitness(certToVerify, witnessAgain, table[0].tree.pubParams) {
		t.Error("Json conversion failed, does not recognize the witness?")
	}
}
func TestJsonConverterNotInTree(t *testing.T) {
	fmt.Println("TestJsonConverter Running")
	certToVerify := testCerts.certs[50123]
	tree := BuildTree(testCerts.certs[:50000], 2, numThreads)
	witness := createWitness(certToVerify, *tree)
	if verifyWitness(certToVerify, witness, tree.pubParams) {
		t.Error("Should not have been in the tree, found it before conversion though.")
	}

	jsooon := genJsonWitness(witness)
	witnessAgain := getWitnessFromJson(jsooon)
	if verifyWitness(certToVerify, witnessAgain, tree.pubParams) {
		t.Error("Should not have been in tree, found it after conversion though.")
	}
}

// go test -run TestSizeOfWitnesses | tee proofSizesMerkle.txt
func TestSizeOfWitnesses(t *testing.T) {
	fmt.Println("TestSizeOfWitnesses Running")

for _, w := range certAmount.c {
	randInt := rand.Intn(w)
	randomcertificate := testCerts.certs[randInt]
	for _, v := range fanOuts.v {
		testTree := BuildTree(testCerts.certs[:w], v, numThreads)
		size := genJsonWitness(createWitness(randomcertificate,*testTree))
		fmt.Println("fan-out: ", v, ", certificates: ", w, ". Witness/membershipProof size in bytes: ", len(size))
	}
	}
}

// Testing for the best amount of threads depending on fanout for the pc building the tree.
// Appears to be equal to the amount of cores on the CPU.
// 8 for Ryzen 7 4700u
func TestDifferentAmountOfThreads(t *testing.T) {
	fmt.Println("TestDifferentAmountOfThreads -  starting")
	fanOuts := []int{2} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}

	for _, fan := range fanOuts {
		for threads := 16; threads < 40; threads++ {
			start := time.Now()
			BuildTree(testCerts.certs, fan, threads)
			elapsed := time.Since(start)
			fmt.Println("Time elapsed making tree with fanout: ", fan, " and threads:", threads, "is: ", elapsed)
		}
	}
}

// Benchmark/party time!!!!!!!
// Benchmark site https://blog.logrocket.com/benchmarking-golang-improve-function-performance/
// Loads all certs, to use less for tests use testCerts.certs[:500] to pick the first 500 certs.
// run benchmarks: go test -bench=benchmarkName -run=^a
// To get memory alloc, run each test 100 times and avoid timeout use:
// go test -bench=BenchmarkBuildTreeTime -run=^a -benchtime=100x -benchmem  -timeout 99999s | tee old.txt

// To run this test
// go test -bench=BenchmarkVerifyNode -run=^a -benchtime=1000x -benchmem  -timeout 99999s | tee merkVerifyNodeBench.txt
func BenchmarkVerifyNode(b *testing.B) {
	fmt.Println("BenchmarkVerifyNode Running")
	b.ResetTimer()
	testAmount := 1000 //Change if you change -benchtime=10000x

	randomCerts := make([][]byte, testAmount)

	for k := range randomCerts {
		randInt := rand.Intn(len(testCerts.certs))
		randomCerts[k] = testCerts.certs[randInt]
		//(0, len(testCerts.certs))
	}

	for _, v := range table {
		b.Run(fmt.Sprintf("fanOut: %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				verifyNode(randomCerts[i], v.tree)
			}
		})

	}
}

// go test -bench=BenchmarkDifferentAmountOfThreads -benchtime=10x -run=^a -benchmem  -timeout 99999s | tee merkBenchmarkDifferentAmountOfThreadsBench.txt
func BenchmarkDifferentAmountOfThreads(b *testing.B) {
	fmt.Println("BenchmarkDifferentAmountOfThreads -  starting")
	fanOuts := []int{2} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}

	for _, fan := range fanOuts {
		for threads := 1; threads < 20; threads++ {
			b.Run(fmt.Sprintf("fanOut: %d and threads %d", fan, threads), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					BuildTree(testCerts.certs, fan, threads)

				}
			})
		}
	}
}

// go test -bench=BenchmarkDifferentAmountOfCertsBuild -benchtime=10x -run=^a -benchmem  -timeout 99999s | tee merkBenchmarkDifferentAmountOfCertsBench.txt
func BenchmarkDifferentAmountOfCertsBuild(b *testing.B) {
	fmt.Println("BenchmarkRootDifferentAmountOfCertsBuild -  starting")
	fanOuts := []int{2} //, 4, 8, 16, 32, 64, 128, 256, 512, 1024}

	for _, fan := range fanOuts {
		for _, amountOfCerts := range certAmount.c {
			b.Run(fmt.Sprintf("fanOut: %d and amountOfCerts %d", fan, amountOfCerts), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					fmt.Println("len of certs", len(testCerts.certs))
					BuildTree(testCerts.certs[:amountOfCerts], fan, numThreads)

				}
			})
		}
	}
}

// TODO test on server
// go test -bench=BenchmarkBuildTreeTime -run=^a -benchtime=10x -benchmem  -timeout 999999s | tee merkBuildTreeBench.txt
func BenchmarkBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkBuildTreeTime - starting")
	b.ResetTimer()
	for _, certs := range certAmount.c {
		for _, f := range fanOuts.v { //Different fanouts
			b.Run(fmt.Sprintf("fan-out: %d, certs: %d", f, certs), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					BuildTree(testCerts.certs[:certs], f, numThreads)
				}
			})
		}
	}
}

// TODO run benchmark on server, remember to divide result by 1000!!!
// go test -bench=BenchmarkCreateWitness -run=^a -benchtime=200x -benchmem  -timeout 99999s | tee merkBenchmarkCreateWitness.txt
func BenchmarkCreateWitness(b *testing.B) {
	fmt.Println("BenchmarkCreateWitness Running")
	testAmount := 200 //Change if you change -benchtime=10000x
	randomCerts := make([][]byte, testAmount)

	for _, certs := range certAmount.c {
		for k := range randomCerts {
			randInt := rand.Intn(len(testCerts.certs))
			randomCerts[k] = testCerts.certs[randInt]
		}
		for _, f := range fanOuts.v {
			benchTree := BuildTree(testCerts.certs[:certs], f, numThreads)
			b.ResetTimer()
			b.Run(fmt.Sprintf("fanOut: %d, certs: %d", f, certs), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					createWitness(randomCerts[i], *benchTree)

				}
			})
		}
	}
}

// TODO run benchmark on server
// sudo go test -bench=BenchmarkVerifyWitness -run=^a -benchtime=1000x -benchmem  -timeout 99999s | sudo tee merkBenchmarkVerifyWitnessV2.txt
func BenchmarkVerifyWitness(b *testing.B) {
	fmt.Println("BenchmarkVerifyWitness - starting")
	testAmount := 1000 //Change if you change -benchtime=1000x
	certsToTest := make([][]byte, testAmount)
	witnesses := make([]witness, testAmount)

	//Get certs to test

	b.ResetTimer()
	for _, certs := range certAmount.c {
		for _, f := range fanOuts.v {
			benchTree := BuildTree(testCerts.certs[:certs], f, numThreads)
			// Get random certs, and get witnesses for them in the generated tree.
			for k := range testAmount {
				randInt := rand.Intn(certs)
				certsToTest[k] = testCerts.certs[randInt]
				witnesses[k] = createWitness(certsToTest[k], *benchTree)
			}
			b.Run(fmt.Sprintf("fan-out: %d, certs: %d", f, certs), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					verifyWitness(certsToTest[i], witnesses[i], benchTree.pubParams)
				}
			})
		}
	}
}
