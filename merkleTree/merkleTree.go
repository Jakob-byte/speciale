package main

import (
	//"bytes"
	"crypto/sha256"
	//"errors"
	"fmt"
	//"hash"
	//"math/big"
	"reflect"
)

type node struct{
	parent *node
	children []*node
	ownHash [32]byte
	leaf bool
	certificate []byte
	duplicate bool
}

type merkleTree struct{
	Root *node
	leafs []*node
	fanOut int
}


func BuildTree(certs []string, fanOut int) *merkleTree {
	var merk merkleTree

	//var leafs []*node
	

	uneven := false
	if len(certs) % 2 == 1 {
		certs = append(certs, certs[len(certs)])
		uneven = true
	} 

	leafs := make([]*node, len(certs))
	
	for i := 0; i < len(certs); i++ {
		byteCet := []byte(certs[i])
		testHash := sha256.Sum256(byteCet)
		leafs[i]= &node { 
			certificate: byteCet,
			ownHash: testHash,
			leaf: true,
			duplicate: false,
		}
	}

	if uneven {
		leafs[len(leafs)].duplicate = true
	}

	fmt.Println(reflect.TypeOf(leafs))

	nextLayer := makeLayer(leafs)
	fmt.Println(len(nextLayer))
	for len(nextLayer) > 1{
		nextLayer = makeLayer(nextLayer)
		fmt.Println(len(nextLayer))
	}
	
	merk = merkleTree {
		fanOut: fanOut,
		Root: nextLayer[0],
		leafs: leafs,
	}

	return &merk
}

func makeLayer (nodes []*node) []*node {

	nextLayer := make([]*node, len(nodes)/2) // divided with fanout which is 2 in this case

	for i := 0; i < len(nodes);{
		n1 := nodes[i]
		n2 := nodes[i+1]
		fmt.Println(n2.ownHash[:])
		sum := sha256.Sum256(append(n1.ownHash[:], n2.ownHash[:]...))
		nextLayer[i/2] = &node {
			ownHash: sum,
			leaf: false,
			children: []*node {n1, n2},
		}
		n1.parent = nextLayer[i/2]
		n2.parent = nextLayer[i/2]
		i = i+2
	}
	
	return nextLayer
}

func main(){
	l := []string {"gedshsfhdfghfghd", "jfdghfghfdghfghens", "dortfghfdghfdghfdgjhe", "fledfhfgjfdjdfgjdfghmming"}
	merkTree := BuildTree(l, 2)
	fmt.Println(merkTree)
	fmt.Println("Succes")
}