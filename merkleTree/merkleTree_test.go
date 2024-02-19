package main

import "testing"

func TestVerifyTree(t *testing.T) {
	result := true
	certArray := loadCertificates("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result = verifyTree(certArray, *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}

func TestVerifyCert(t *testing.T) {
	certArray := loadCertificates("testCerts/")
	merkTree := BuildTree(certArray, 2)
	result := verifyNode(certArray[5], *merkTree)
	if result != true {
		t.Errorf("Result was incorrect, got: %t, want: %t.", result, true)
	}
}
