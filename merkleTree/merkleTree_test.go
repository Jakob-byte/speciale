package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestVerifyTree(t *testing.T) {
	fmt.Println("TestVerifyTree -  starting")
	certArray := loadCertificatesFromOneFile("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result := verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	fmt.Println("TestVerifyCert -  starting")

	certArray := loadCertificatesFromOneFile("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result := verifyNode(certArray[5], *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestTreeBuilder(t *testing.T) {
	fmt.Println("TestTreeBuilder -  starting")
	max := 1000
	min := 10
	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max-min) + min
		certArray := loadCertificatesFromOneFile("testCerts/", randNumb)
		merkTree := BuildTree(certArray, 2)
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

func TestDifferentFanOuts(t *testing.T) {
	fmt.Println("TestDifferentFanOuts -  starting")
	max := 500
	min := 100
	maxFan := 100
	minFan := 2
	for i := 0; i < 10; i++ {
		randNumb := rand.Intn(max-min) + min
		fanNumb := rand.Intn(maxFan-minFan) + minFan
		certArray := loadCertificatesFromOneFile("testCerts/", randNumb)
		merkTree := BuildTree(certArray, fanNumb)
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
	certArray := loadCertificatesFromOneFile("testCerts/")
	merkTree := BuildTree(certArray, 2)
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
	certArray := loadCertificatesFromOneFile("testCerts/")
	merkTree := BuildTree(certArray, 2)
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
// Let's GOOOO!!!!!!!!!!!!!!!!!

var table = []struct {
	fanOut int
	tree   merkleTree
}{
	//{input: 1}, Doesn't work for some reasone :D
	{fanOut: 2, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 2)},
	{fanOut: 3, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 3)},
	{fanOut: 4, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 4)},
	{fanOut: 5, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 5)},
	{fanOut: 6, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 6)},
	{fanOut: 7, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 7)},
	{fanOut: 8, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 8)},
	{fanOut: 9, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 9)},
	{fanOut: 10, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 10)},
	{fanOut: 11, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 11)},
	{fanOut: 12, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 12)},
	{fanOut: 13, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 13)},
	{fanOut: 14, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 14)},
	{fanOut: 15, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 15)},
	{fanOut: 16, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 16)},
	{fanOut: 17, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 17)},
	{fanOut: 18, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 18)},
	{fanOut: 19, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 19)},
	{fanOut: 20, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 20)},
	//{fanOut: 25, tree: *BuildTree(loadCertificatesFromOneFile("banan"), 25)},
}

func BenchmarkBuildTreeTime(b *testing.B) {
	fmt.Println("BenchmarkBuildTreeTime Running")
	certArray := loadCertificatesFromOneFile("testCerts/")
	b.ResetTimer()
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BuildTree(certArray, v.fanOut)

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
	testCerts := loadCertificatesFromOneFile("testCerts/")
	b.ResetTimer()

	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			//b.ResetTimer()
			for i := 0; i < b.N; i++ {
				verifyNode(testCerts[i], v.tree)
			}
		})
	}
}
