package verkletree

import (
	"crypto/rand"

	//"fmt"
	//"runtime"
	"slices"
	"sync"

	//combin "gonum.org/v1/gonum/stat/combin"
	//	"math/big"
	e "github.com/cloudflare/circl/ecc/bls12381"
)

// The public key struct which contains the g1's, alpha^i*g1. As well as g2 and alpha*g2
type PK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
}

// The struct which contains the polynomial stored as a slice.
type poly struct {
	coefficients []e.Scalar
}

// The witness struct which contains the necessary info for a witness to prove it is contained in a commitment.
type witnessStruct struct {
	Index uint64
	Fx0   e.Scalar
	W     e.G1
}

var mutexBuddy sync.Mutex

// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
// The setup function handles det setup of the crypto part of the the VerkleTree with the elliptic curves and fields, takes as input a security parameter.
// It returns the public key.
func setup(security, t int) PK {
	//Sets up the generator elements, as well as the secret key a.
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	//Makes the list which containts g_i*alpha^i
	gList := make([]e.G1, t)
	at := new(e.Scalar)
	at.SetString("1")
	for i := 0; i < t; i++ {
		gList[i].ScalarMult(at, g1)
		//fmt.Println(gList[i].String())
		at.Mul(at, a)
	}

	ag2 := new(e.G2)
	ag2.ScalarMult(a, g2)

	//Creates the public key
	pk := PK{
		g1:       *g1,
		alphaG1s: gList,
		g2:       *g2,
		alphaG2:  *ag2,
	}
	return pk
}

// Changes the certificates (bytes) to scalars and returns it as a list.
func certToScalarVector(certs [][]byte) []e.Scalar {
	vects := make([]e.Scalar, len(certs))
	for i, v := range certs {
		vects[i].SetBytes(v)
	}
	return vects
}

// Evaluates the polynomium for a given x
func calcPoly(x uint64, poly poly) e.Scalar {
	var answer e.Scalar
	var ansToBe e.Scalar
	var xScalar e.Scalar
	xScalar.SetUint64(x)
	for i, v := range poly.coefficients {

		ansToBe.SetOne()
		// Does x^i, as e.Scalar does not have a exp/pow function
		for j := 0; j < i; j++ {
			ansToBe.Mul(&ansToBe, &xScalar)
		}
		ansToBe.Mul(&v, &ansToBe)
		answer.Add(&answer, &ansToBe)

	}
	return answer
}

// Calculates the divisor used in buildtree. Takes the fanout and the unique combinations as input.
// Returns the divisors as a list.
// TODO redegør for det her med math!
func divisorCalculator(fanOut int, degreeComb [][][]int) []e.Scalar {
	divisorList := make([]e.Scalar, fanOut)
	var divisor e.Scalar
	divisor.SetOne()
	var iScalar e.Scalar
	for i := 0; i < fanOut; i++ {
		if i != 0 {
			iScalar.SetUint64(uint64(i))
			divisor.Mul(&divisor, &iScalar)
		}
	}
	divisor.Inv(&divisor)
	divisorList[0] = divisor

	var divisorMinusI e.Scalar
	var divToBe e.Scalar
	var sumDiv e.Scalar
	var cScalar e.Scalar
	var iInPowerOfJ e.Scalar
	//TODO black magic
	for i := 1; i < fanOut; i++ {
		divisorMinusI.SetUint64(0)
		for j, combs := range degreeComb {
			sumDiv.SetUint64(0)
			for _, comb := range combs {
				if !slices.Contains(comb, 0) && !slices.Contains(comb, i) {
					divToBe.SetOne()
					for _, c := range comb {
						cScalar.SetUint64(uint64(c))
						divToBe.Mul(&divToBe, &cScalar)
					}
					iInPowerOfJ.SetOne()
					iScalar.SetUint64(uint64(i))
					for k := 1; k <= j+1; k++ {
						iInPowerOfJ.Mul(&iInPowerOfJ, &iScalar)
					}
					divToBe.Mul(&divToBe, &iInPowerOfJ)

					sumDiv.Add(&sumDiv, &divToBe)

				}
			}
			if ((j) % 2) == 0 {
				//sumDiv *= -1
				sumDiv.Neg()
			}
			//divisorMinusI += sumDiv
			divisorMinusI.Add(&divisorMinusI, &sumDiv)
		}
		// inverse it so when we multiply with it, it will work as division!!!

		divisorMinusI.Inv(&divisorMinusI)

		divisorList[i] = divisorMinusI
	}
	return divisorList
}

func lagrangeBasisForGivenI(indexI int, fanOut int, divisorList []e.Scalar, degreeComb [][][]int, lagrangeBasisList *[][]e.Scalar) []e.Scalar {
	var coefToBe e.Scalar
	var combScalar e.Scalar
	divisorMinusI := divisorList[indexI]
	coefToBeList := make([]e.Scalar, fanOut-1)

	// The loop starts by looking at the first length of unique combinations. E.g. combinations of 0, 1, 2, 3, 4, 5. Then the next will be 0, 1, 2, 3, 4 and so on.
	for j, combs := range degreeComb {
		coefToBeList[j].SetUint64(0)
		// It then looks at one of these unique combinations e.g. 5,0,2,3,1,5
		for _, comb := range combs {
			// If the slice (unique combination) contains either 0 or the index we're looking at (from the input vector) we skip it.
			if !slices.Contains(comb, 0) && !slices.Contains(comb, indexI) {
				//We then go through the slice, and multiply the values together as scalars.
				coefToBe.SetOne()
				for _, c := range comb {
					combScalar.SetUint64(uint64(c))
					coefToBe.Mul(&coefToBe, &combScalar)
				}
				//If the comb length (j) is even we negate coefToBe,
				//before multiplying it with with the value from our input vector
				if ((j) % 2) == 0 {
					coefToBe.Neg()
				}
				coefToBe.Mul(&coefToBe, &divisorMinusI)

				coefToBeList[j].Add(&coefToBe, &coefToBeList[j])
			}

		}
	}
	mutexBuddy.Lock()
	defer mutexBuddy.Unlock()
	(*lagrangeBasisList)[indexI] = coefToBeList
	return coefToBeList
}

func lagrangeBasisCalc(fanOut int, degreeComb [][][]int, divisorList []e.Scalar) [][]e.Scalar {
	// var lagrangeBasisList [][]e.Scalar
	lagrangeBasisList := make([][]e.Scalar, fanOut)
	var wg sync.WaitGroup
	for i := 0; i < fanOut; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			lagrangeBasisForGivenI(index, fanOut, divisorList, degreeComb, &lagrangeBasisList)
		}(i)
		//numGoroutines := runtime.NumGoroutine()
		//fmt.Println("Number of active goroutines:", numGoroutines)

		//lagrangeBasisList = append(lagrangeBasisList, lagrangeBasisForGivenI(i, fanOut, divisorList, degreeComb, &lagrangeBasisList))
	}
	wg.Wait()
	//fmt.Println("LAGRANGEBASISLIST:", lagrangeBasisList)
	return lagrangeBasisList
}

// This translates the input vector into a polynomial which can be used for KZG commitment. It takes the scalar vector as input, unique combinations and divisorlist.
// It returns the polynomial of the vector, f(i)=scalarVect[i].
func realVectorToPoly(scalarVect []e.Scalar, lagrangeBasisList [][]e.Scalar) poly {
	//Prepares variable for the polynomial.
	var answer poly
	coefs := make([]e.Scalar, len(scalarVect))
	coefs[0] = scalarVect[0] // first value in list of points, this is a constant coefficient in the polynomial (aka the first coefficient if a0 + a1x + a2x^2 + ...)
	var coefToBe e.Scalar
	//lagrangeBasisList := lagrangeBasisCalc(len(scalarVect), degreeComb, divisorList)
	for i, y := range scalarVect {
		for j, eScalar := range lagrangeBasisList[i] {
			coefToBe = eScalar

			coefToBe.Mul(&coefToBe, &y)
			coefs[j+1].Add(&coefs[j+1], &coefToBe)
		}
	}
	answer.coefficients = coefs
	return answer
}

// Takes the polynomial and x0 as input, and returns the quotient as a polynomial struct.
func quotientOfPoly(polynomial poly, x0 uint64) poly {
	var quotient poly
	degree := len(polynomial.coefficients)
	coefs := make([]e.Scalar, degree-1)
	var xPower e.Scalar
	var xNul e.Scalar
	var mulSomething e.Scalar

	xNul.SetUint64(x0)
	//TODO Coeffients of the Quotient computed as coef 1 = a1*x0^0+a2*x0^1 do the math and put in report or something
	for i := range polynomial.coefficients[1:] { //we ignore the forst coeff as it is divided out
		xPower.SetOne()
		for j := i; j < len(coefs); j++ {
			mulSomething.Mul(&polynomial.coefficients[j+1], &xPower)
			coefs[i].Add(&coefs[i], &mulSomething)
			xPower.Mul(&xPower, &xNul)
		}
	}
	quotient.coefficients = coefs
	return quotient
}

// takes as input the certifictes as a list of bytes, and the lagrangeBasis and returns the lagrange polynomial for certificates and lagrangebasis
// the lagrangebasisLsit and the certVect must have same length
func certVectorToPolynomial(certVect [][]byte, lagrangeBasisList [][]e.Scalar) poly {

	scalarVector := certToScalarVector(certVect)

	polynomial := realVectorToPoly(scalarVector, lagrangeBasisList)
	return polynomial
}

// Commit function, that computes the KZG polynomial commitment, given the public key and polynomial
func commit(pk PK, polynomial poly) e.G1 {
	var commitment e.G1
	var cToBe e.G1
	// computes the commitment as coef_i*(a^i*g_1) for i=0 to degree of polynomial
	for i, coef := range polynomial.coefficients {
		cToBe.ScalarMult(&coef, &pk.alphaG1s[i])
		if i == 0 {
			commitment = cToBe
		} else {
			commitment.Add(&cToBe, &commitment) //
		}
	}
	return commitment
}

func open() int { //TODO fiks den aka. lav den
	return 0
}

// verifies that the polynomial for the commitment is correct, by recomputing the commitment for the polynomial and checking it is the same as the one to verify
func verifyPoly(pk PK, commitmentToVerify e.G1, polynomial poly) bool {
	commitment := commit(pk, polynomial)

	return commitment.IsEqual(&commitmentToVerify)
}

// Creates the witness for the specified index of the polynomial
// computing the quotientPolynomial for the given index and then calculating a commitment for the quotient and a evaluation of the index of the original polynomial
func createWitness(pk PK, polynomial poly, index uint64) witnessStruct {
	//HokusPokusDinKatErIFokus()
	quotientPoly := quotientOfPoly(polynomial, index)
	w := commit(pk, quotientPoly)
	fx0 := calcPoly(index, polynomial)
	witness := witnessStruct{
		W:     w,
		Index: index,
		Fx0:   fx0,
	}
	return witness
}

// Verifies the witness corresponds with the commitment
func verifyWitness(pk PK, commitment e.G1, witness witnessStruct) bool {
	lSide := e.Pair(&commitment, &pk.g2)

	// e(w, alpha * g2 - x0 * g2) * e(g1, g2) ^f(x_0)
	var alphaG2minusX0G2 e.G2
	var x0 e.Scalar
	x0.SetUint64(uint64(witness.Index))
	var x0g2 e.G2
	x0g2.ScalarMult(&x0, &pk.g2)
	x0g2.Neg()
	alphaG2minusX0G2.Add(&pk.alphaG2, &x0g2) //
	rSide1 := e.Pair(&witness.W, &alphaG2minusX0G2)
	rSide2 := e.Pair(&pk.g1, &pk.g2)

	// try the other Pair Function pairPRod first make into list
	rSide2.Exp(rSide2, &witness.Fx0)

	rSide1.Mul(rSide1, rSide2)

	return lSide.IsEqual(rSide1)
}
