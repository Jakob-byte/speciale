package verkletree

import (
	"fmt"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

// Calculates and returns x^n
func powerSimple(x e.Scalar, n int) e.Scalar {
	var result e.Scalar
	result.SetOne()
	for j := 0; j < n; j++ {
		result.Mul(&result, &x)
	}
	fmt.Println("Hej")
	return result
}
