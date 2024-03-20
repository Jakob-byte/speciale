package main

import (
	"fmt"
	"math/rand"
	"testing"
	//"time"
)

func TestVerifyTree(t *testing.T) {
	fmt.Println("TestVerifyTree -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	merkTree := BuildTree(certArray, 2,500)
	result := verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	fmt.Println("TestVerifyCert -  starting")

	certArray := loadCertificates("AllCertsOneFile20000", 1)

	
	fmt.Println("len of certarray", len(certArray))
	merkTree := BuildTree(certArray, 2,500)
	//for i:= 0 ; i<10; i++ {
	//	fmt.Println(i, "hash at index , merkTree.leafs[i].parent.parent.ownHash)
	//}
	result := verifyNode(certArray[2], *merkTree)
	
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
		fmt.Println(certArray[3204043959346])
		
		
	}

}


func TestTreeBuild2(t *testing.T) {
	fmt.Println("TestTreeBuild -  starting")
	certArray := loadCertificates("AllCertsOneFile20000", 5)
	BuildTree(certArray, 15,500)
}

func TestTreeBuilder(t *testing.T) {
	fmt.Println("TestTreeBuilder -  starting")
	max := 1000
	min := 10
	for i := 9; i < 10; i++ {
		randNumb := rand.Intn(max-min) + min
		certArray := loadCertificates("AllCertsOneFile20000", 2)
		merkTree := BuildTree(certArray, 2,500)
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
	for i := 1; i < 2; i++ {
		randNumb := rand.Intn(max-min) + min
		fanNumb := rand.Intn(maxFan-minFan) + minFan
		certArray := loadCertificates("AllCertsOneFile20000", 2)
		merkTree := BuildTree(certArray, fanNumb,500)
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
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	merkTree := BuildTree(certArray, 2,500)
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
	certArray := loadCertificates("AllCertsOneFile20000", 2)
	merkTree := BuildTree(certArray, 2,500)
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
var testCerts = struct {
	certs [][]byte
}{
	//certs: loadCertificates("AllCertsOneFIle20000", 5),
}

var table = []struct {
	fanOut int
	tree   merkleTree
}{
	//{input: 1}, Doesn't work for some reasone :D
	//{fanOut: 2, tree: *BuildTree(testCerts.certs, 2)},
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
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size %d", v.fanOut), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BuildTree(testCerts.certs, v.fanOut,500)

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
