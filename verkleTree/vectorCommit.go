package verkletree

import (
	"crypto/rand"

	"slices"
	//combin "gonum.org/v1/gonum/stat/combin"
	//	"math/big"
	e "github.com/cloudflare/circl/ecc/bls12381"
)

//The public key struct which contains the g1's, alpha^i*g1. As well as g2 and alpha*g2
type PK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
}

//The struct which contains the polynomial stored as a slice.
type poly struct {
	coefficients []e.Scalar
}

//The witness struct which contains the necessary info for a witness to prove it is contained in a commitment.
type witnessStruct struct{
	index uint64
	fx0 e.Scalar
	w e.G1
}

// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
//The setup function handles det setup of the crypto part of the the VerkleTree with the elliptic curves and fields, takes as input a security parameter . 
//It returns the public key.
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
	pk := PK{
		g1:       *g1,
		alphaG1s: gList,
		g2:       *g2,
		alphaG2:  *ag2,
	}
	return pk
}

func certToScalarVector(certs [][]byte) []e.Scalar {
	vects := make([]e.Scalar, len(certs))
	for i, v := range certs {
		vects[i].SetBytes(v)
	}
	return vects
}

func calcPoly(x uint64, poly poly) e.Scalar {
	var answer e.Scalar
	var ansToBe e.Scalar
	var xScalar e.Scalar
	xScalar.SetUint64(x)
	for i, v := range poly.coefficients {

		ansToBe.SetOne()
		for j := 0; j < i; j++ {
			ansToBe.Mul(&ansToBe, &xScalar)
		}
		ansToBe.Mul(&v, &ansToBe)
		answer.Add(&answer, &ansToBe)

	}
	return answer
}

func dividentCalculator (fanOut int, degreeComb [][][]int) []e.Scalar{
	dividentList := make([]e.Scalar,fanOut)
	var divident e.Scalar
	divident.SetOne()
	var iScalar e.Scalar
	for i := 0; i<fanOut; i++ {
		if i != 0 {
			iScalar.SetUint64(uint64(i))
			divident.Mul(&divident, &iScalar)
		}
	}
	divident.Inv(&divident)
	dividentList[0] = divident
	
	var dividentMinusI e.Scalar
	var divToBe e.Scalar
	var sumDiv e.Scalar
	var cScalar e.Scalar
	var iInPowerOfJ e.Scalar
	for i:=1; i<fanOut; i++ {
		dividentMinusI.SetUint64(0)
			for j, combs := range degreeComb {
				sumDiv.SetUint64(0)
				for _, comb := range combs {
					if  !slices.Contains(comb, 0) && !slices.Contains(comb, i) {
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
				//dividentMinusI += sumDiv
				dividentMinusI.Add(&dividentMinusI, &sumDiv)
			}
		// inverse it so when we multiply with it, it will work as division!!!

		dividentMinusI.Inv(&dividentMinusI)

		dividentList[i] = dividentMinusI
	}
	return dividentList
}


func realVectorToPoly(points []e.Scalar, degreeComb [][][]int, dividentList []e.Scalar) poly {
	var answer poly
	coefs := make([]e.Scalar, len(points))
	coefs[0] = points[0] // first value in list of points, this is constant coefficient

	for i, y := range points {
	
		dividentMinusI := dividentList[i]
		
		var coefToBe e.Scalar	
		var combScalar e.Scalar
		for j, combs := range degreeComb {
			for _, comb := range combs {
				if !slices.Contains(comb, 0) && !slices.Contains(comb, i) {
					coefToBe.SetOne()
					for _, c := range comb {
						combScalar.SetUint64(uint64(c))
						coefToBe.Mul(&coefToBe, &combScalar)
					}
					if ((j) % 2) == 0 {
						// Here order - coef
						coefToBe.Neg()
					}
					coefToBe.Mul(&coefToBe, &y)

					coefToBe.Mul(&coefToBe, &dividentMinusI)
					coefs[j+1].Add(&coefs[j+1], &coefToBe)
				}
			}
		}

	}
	//fmt.Println("coefs", coefs)
	answer.coefficients = coefs
	return answer
}

func quotientOfPoly(polynomial poly, x0 uint64) poly {
	var quotient poly
	degree := len(polynomial.coefficients)
	coefs := make([]e.Scalar, degree-1)
	var xPower e.Scalar
	var xNul e.Scalar
	var mulSomething e.Scalar

	xNul.SetUint64(x0)
	//fmt.Println("coefficients len: ", len(polynomial.coefficients))
	for i := range polynomial.coefficients[1:] { //we ignore the forst coeff as it is divided out
		xPower.SetOne()
		for j := i; j < len(coefs); j++ {
			//fmt.Println("j:", j)
			//coefs[i] += polynomial.coefficients[j+1] * math.Pow(x0, float64(count))
			mulSomething.Mul(&polynomial.coefficients[j+1], &xPower)
			coefs[i].Add(&coefs[i], &mulSomething)
			xPower.Mul(&xPower, &xNul)
			//fmt.Println("OG coefs: ", polynomial.coefficients[j+1])
			//fmt.Println("coefs[i]=v", i, coefs[i])
		}
	}
	quotient.coefficients = coefs
	return quotient
}

func certVectorToPolynomial(certVect [][]byte, degreeComb [][][]int, dividentList []e.Scalar) poly {
	scalarVector := certToScalarVector(certVect)
	polynomial := realVectorToPoly(scalarVector, degreeComb,dividentList)
	return polynomial
}

func commit(pk PK, polynomial poly) e.G1 {
	var commitment e.G1
	var cToBe e.G1
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

func verifyPoly(pk PK, commitmentToVerify e.G1, polynomial poly) bool {
	commitment := commit(pk, polynomial)

	return commitment.IsEqual(&commitmentToVerify)
}

func createWitness(pk PK, polynomial poly, index uint64) witnessStruct {
	//HokusPokusDinKatErIFokus()
	quotientPoly := quotientOfPoly(polynomial, index)
	w := commit(pk, quotientPoly)
	fx0 := calcPoly(index, polynomial) 
	witness := witnessStruct {
		w: w,
		index: index,
		fx0: fx0,
	}

	return witness
}

func verifyWitness(pk PK, commitment e.G1, witness witnessStruct) bool {
	lSide := e.Pair(&commitment, &pk.g2)

	// e(w, alpha * g2 - x0 * g2) * e(g1, g2) ^f(x_0)
	var alphaG2minusX0G2 e.G2
	var x0 e.Scalar
	x0.SetUint64(uint64(witness.index))
	var x0g2 e.G2
	x0g2.ScalarMult(&x0, &pk.g2)
	x0g2.Neg()
	alphaG2minusX0G2.Add(&pk.alphaG2, &x0g2) //
	rSide1 := e.Pair(&witness.w, &alphaG2minusX0G2)
	rSide2 := e.Pair(&pk.g1, &pk.g2)

	// try the other Pair Function pairPRod first make into list
	rSide2.Exp(rSide2, &witness.fx0)

	rSide1.Mul(rSide1, rSide2)

	return lSide.IsEqual(rSide1)
}
