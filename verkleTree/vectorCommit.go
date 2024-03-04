package verkletree

import (
	"crypto/rand"
	"math"
	"math/big"
	"slices"

	combin "gonum.org/v1/gonum/stat/combin"

	//	"math/big"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

type PK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
}

type poly struct {
	coefficients []e.Scalar
}

func certToScalarVector(certs [][]byte) []e.Scalar {
	vects := make([]e.Scalar, len(certs))
	for i, v := range certs {
		vects[i].SetBytes(v)
	}
	return vects
}
func calcPoly(x e.Scalar, poly poly) e.Scalar {
	var answer e.Scalar
	for i, v := range poly.coefficients {
		ansToBe := x
		for j := 0; j < i; j++ {
			ansToBe.Mul(&ansToBe, &x)
		}
		ansToBe.Mul(&v, &ansToBe)
		answer.Add(&answer, &ansToBe)
		//answer.Add(answer, a*math.Pow(x, float64(i)))
	}
	return answer
}

func realVectorToPoly(points []e.Scalar) poly {
	var answer poly
	coefs := make([]e.Scalar, len(points))
	coefs[0] = points[0] // first value in list of points, this is constant coefficient
	divident := 1.0

	//divident Finder loop
	for i := range points {
		if i != 0 {
			divident = divident * float64(i)
		}
	}
	var degreeComb [][][]int
	for k := len(points) - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(len(points), k-1))
	}
	//flipBool := true
	var dividentMinusI float64
	var divToBe float64
	var sumDiv float64
	for i, y := range points {
		dividentMinusI = 0
		if i == 0 {
			dividentMinusI = divident
		} else {

			for j, combs := range degreeComb {
				sumDiv = 0
				for _, comb := range combs {
					if !slices.Contains(comb, i) {
						divToBe = 1.0
						for _, c := range comb {
							divToBe *= float64(c)
						}
						divToBe *= math.Pow(float64(i), float64(j+1))

						sumDiv += divToBe
					}

				}

				if ((j) % 2) == 0 {
					sumDiv *= -1
				}
				dividentMinusI += sumDiv
			}

		}
		//WE can reuse the math for what to divide with! Convert the float thingie to bytes!!!
		var dividentScalar e.Scalar
		if dividentMinusI < 0 {
			dividentMinusI *= -1
			dividentScalar.SetUint64(uint64(dividentMinusI))
			dividentScalar.Neg()
		} else {
			dividentScalar.SetUint64(uint64(dividentMinusI))
		}
		var coefToBe e.Scalar
		var combScalar e.Scalar
		for j, combs := range degreeComb {

			for _, comb := range combs {
				if !slices.Contains(comb, i) {
					coefToBe.SetOne()
					for _, c := range comb {
						combScalar.SetUint64(uint64(c))
						coefToBe.Mul(&coefToBe, &combScalar)
					}

					if ((j) % 2) == 0 {
						coefToBe.Neg()
					}
					coefToBe.Mul(&coefToBe, &y)

					dividentScalar.Inv(&dividentScalar)
					coefs[j+1].Mul(&coefToBe, &dividentScalar)
					//coefs[j+i] += (coefToBe * y) / dividentMinusI
				}
			}
		}

	}
	//fmt.Println("coefs", coefs)
	answer.coefficients = coefs
	return answer
}

// what bit security do we have or bls12381
// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
func setup(security int, t int) PK {
	//fmt.Println("42 is the answer")
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	gList := make([]e.G1, t)
	at := new(e.Scalar)
	at.Set(a)
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

// TODO CHANGE VECTTOCOMMIT TO POLYNOMIAL!!!!!
func commit(pk PK, vectToCommit [][]byte) e.G1 {
	var commitment e.G1
	polyCoefs := vectToPolySike(vectToCommit) //TODO make a poly before calling commit.
	//	polyCoefs := vectToPoly(input)
	phiScalar := new(e.Scalar)
	var cToBe e.G1
	for i, phi := range polyCoefs {
		phiScalar.SetBytes(phi.Bytes())
		cToBe.ScalarMult(phiScalar, &pk.alphaG1s[i])
		commitment.Add(&cToBe, &commitment) //TODO Should there be a "mult" here somehow, as that is what is described in the original KZG paper.
	}

	return commitment
}

func open() int { //TODO fiks den aka. lav den
	return 0
}

func verifyPoly(pk PK, commitmentToVerify e.G1, vectToCommit [][]byte) bool {
	commitment := commit(pk, vectToCommit)
	return commitment.IsEqual(&commitmentToVerify)
}

func createWitness(pk PK, vectToCommit [][]byte, index int) (int, []byte, e.G1) {

	phiI := vectToCommit[index] //TODO to call to polynomial, when we get that to work
	w := commit(pk, vectToCommit)
	return index, phiI, w
}

func verifyEval(pk PK, commitment e.G1, index int, vectToCommit [][]byte, witness e.G1) bool {
	lSide := e.Pair(&commitment, &pk.g2)

	// e(w, alpha * g2 - x0 * g2) * e(g1, g2) ^f(x_i)
	var alphaG2minusX0G2 e.G2
	xi := new(e.Scalar)
	xi.SetUint64(uint64(index))
	var xig2 e.G2
	xig2.ScalarMult(xi, &pk.g2)
	alphaG2minusX0G2.Add(&pk.alphaG2, &xig2)
	rSide1 := e.Pair(&witness, &alphaG2minusX0G2)
	rSide2 := e.Pair(&pk.g1, &pk.g2)
	fxi := new(e.Scalar)
	fxi.SetBytes(vectToCommit[index])
	rSide2.Exp(rSide2, fxi)
	rSide1.Mul(rSide1, rSide2)
	return lSide.IsEqual(rSide1)
}

func vectToPolySike(input [][]byte) []big.Int {
	var result []big.Int
	for _, v := range input {
		var intBig big.Int
		result = append(result, *intBig.SetBytes(v))
	}
	return result
}
