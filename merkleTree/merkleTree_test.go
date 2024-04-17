package main

import (
	"fmt"
	"math/rand"
	"testing"
	//"time"
)

func TestLoadFunc(t *testing.T) {
	fmt.Println("TestLoadFunc -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 20000)
	if len(certArray) != 20000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 20000)
	}

	certArray = loadCertificates("AllCertsOneFile20000", 40000)
	if len(certArray) != 40000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 40000)
	}

	certArray = loadCertificates("AllCertsOneFile20000", 50000)
	if len(certArray) != 50000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 50000)
	}

	certArray = loadCertificates("AllCertsOneFile20000", 60000)
	if len(certArray) != 60000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 60000)
	}

	certArray = loadCertificates("AllCertsOneFile20000", 1000000)
	if len(certArray) != 1000000 {
		t.Errorf("Result was incorrect, got: %v, want: %v.", len(certArray), 1000000)
	}
}

func TestVerifyTree(t *testing.T) {
	fmt.Println("TestVerifyTree -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 20000)
	merkTree := BuildTree(certArray, 2, 500)
	result := verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	fmt.Println("TestVerifyCert -  starting")

	certArray := loadCertificates("AllCertsOneFile20000", 20000)
	// 1590 certificate
	fmt.Println("len of certarray", len(certArray))
	merkTree := BuildTree(certArray, 2, 2)
	//for i:= 0 ; i<10; i++ {
	//	fmt.Println(i, "hash at index , merkTree.leafs[i].parent.parent.ownHash)
	//}
	result := verifyNode(certArray[2], *merkTree)

	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		//fmt.Println(certArray[3204043959346])
	}

}

func TestTreeBuild2(t *testing.T) {
	fmt.Println("TestTreeBuild -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 5)
	BuildTree(certArray, 15, 500)
}

func TestTreeBuilder(t *testing.T) {
	fmt.Println("TestTreeBuilder -  starting")
	max := 100000
	min := 100
	fanMin := 2
	fanMax := 100
	threadMin := 1
	threadMax := 1000
	for i := 0; i < 100; i++ {
		randNumb := rand.Intn(max-min) + min
		randFan := rand.Intn(fanMax-fanMin) + fanMin
		randThread := rand.Intn(threadMax-threadMin) + threadMin
		certArray := loadCertificates("AllCertsOneFile20000", randNumb)
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

// Hej
func TestDifferentFanOuts(t *testing.T) {
	fmt.Println("TestDifferentFanOuts -  starting")
	max := 500
	min := 100
	maxFan := 100
	minFan := 2
	for i := 1; i < 2; i++ {
		randNumb := rand.Intn(max-min) + min
		fanNumb := rand.Intn(maxFan-minFan) + minFan
		certArray := loadCertificates("AllCertsOneFile20000", 20000)
		merkTree := BuildTree(certArray, fanNumb, 500)
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

func TestUpdateLeafVerifyLeaf(t *testing.T) {
	fmt.Println("TestUpdateLeafVerifyLeaf -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 20000)
	merkTree := BuildTree(certArray, 2, 500)
	newCert := loadOneCert("baguetteCert.crt")
	result := verifyNode(newCert, *merkTree)

	if result != false {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, false)
	}

	updatedTree := updateLeaf(certArray[10], *merkTree, newCert)

	result = verifyNode(certArray[10], *updatedTree)
	if result != false {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, false)
	}

	result = verifyNode(newCert, *updatedTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}

}

func TestUpdateLeafVerifyTree(t *testing.T) {
	fmt.Println("TestUpdateLeafVerifyTree -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 20000)
	merkTree := BuildTree(certArray, 2, 500)
	newCert := loadOneCert("baguetteCert.crt")
	updatedTree := updateLeaf(certArray[10], *merkTree, newCert)

	result := verifyTree(certArray, *updatedTree)
	if result != false {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, false)
	}

	certArray[10] = newCert

	result = verifyTree(certArray, *updatedTree)

	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

// Benchmark/party time!!!!!!!
// Benchmark site https://blog.logrocket.com/benchmarking-golang-improve-function-performance/
// Loads all certs, to use less for tests use testCerts.certs[:500] to pick the first 500 certs.
// run benchmarks: go test -bench=benchmarkName -run=^a
// To get memory alloc, run each test 100 times and avoid timeout use:
// go test -bench=BenchmarkBuildTreeTime -run=^a -benchtime=100x -benchmem  -timeout 99999s | tee old.txt
var testCerts = struct {
	certs [][]byte
}{
	certs: loadCertificates("AllCertsOneFIle20000", 100000),
}

var fanOuts = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}
var threads = struct {
	v []int
}{
	v: []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024},
}

var table = []struct {
	fanOut     int
	tree       merkleTree
	testFanout []int
}{
	//{input: 1}, Doesn't work for some reasone :D
	//{fanOut: 2, tree: *BuildTree(testCerts.certs[:500], 2)},
	//{fanOut: 3, tree: *BuildTree(testCerts.certs, 3)},
	//{fanOut: 4, tree: *BuildTree(testCerts.certs, 4)},
	//{fanOut: 5, tree: *BuildTree(testCerts.certs, 5)},
	//{fanOut: 6, tree: *BuildTree(testCerts.certs, 6)},
	//{fanOut: 7, tree: *BuildTree(testCerts.certs, 7)},
	//{fanOut: 8, tree: *BuildTree(testCerts.certs, 8)},
	//{fanOut: 9, tree: *BuildTree(testCerts.certs, 9)},
	//{fanOut: 10, tree: *BuildTree(testCerts.certs, 10)},
	//{fanOut: 11, tree: *BuildTree(testCerts.certs, 11)},
	//{fanOut: 12, tree: *BuildTree(testCerts.certs, 12)},
	//{fanOut: 13, tree: *BuildTree(testCerts.certs, 13)},
	//{fanOut: 14, tree: *BuildTree(testCerts.certs, 14)},
	//{fanOut: 15, tree: *BuildTree(testCerts.certs, 15)},
	//{fanOut: 16, tree: *BuildTree(testCerts.certs, 16)},
	//{fanOut: 17, tree: *BuildTree(testCerts.certs, 17)},
	//{fanOut: 18, tree: *BuildTree(testCerts.certs, 18)},
	//{fanOut: 19, tree: *BuildTree(testCerts.certs, 19)},
	//{fanOut: 20, tree: *BuildTree(testCerts.certs, 20)},
	//{fanOut: 25, tree: *BuildTree(testCerts.certs, 25)},
}

func BenchmarkBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkBuildTreeTime Running")
	b.ResetTimer()
	for _, v := range fanOuts.v {
		for _, o := range threads.v {
			b.Run(fmt.Sprintf("fanOut: %d, threads: %d", v, o), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					BuildTree(testCerts.certs, v, o)
				}
			})
		}
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
