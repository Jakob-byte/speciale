package verkletree

import (
	"crypto/rand"
	"fmt"

	//"fmt"
	//"runtime"
	"slices"
	"sync"

	//combin "gonum.org/v1/gonum/stat/combin"
	//	"math/big"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

// The public key struct which contains the g1's, alpha^i*g1. As well as g2 and alpha*g2
type rootPK struct {
	g1       e.G1
	alphaG1s []e.G1
	g2       e.G2
	alphaG2  e.G2
}

// The struct which contains the polynomial stored as a slice.
type rootpoly struct {
	coefficients []e.Scalar
}

// The witness struct which contains the necessary info for a witness to prove it is contained in a commitment.
type rootwitnessStruct struct {
	Index uint64
	Fx0   e.Scalar
	W     e.G1
}

type precompute struct {
	invsub []e.Scalar
	ta     [][]e.Scalar
	tk     []e.Scalar
}

type pubParams struct {
	degree        int
	lagrangeBasis []e.G1
	diff2         []e.G2
	domain        []e.Scalar
	aPrimeDomainI []e.Scalar
	precalc       *precompute
	zeroG1        e.Scalar
	oneG1         e.Scalar
}

var rootmutexBuddy sync.Mutex

// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
// The setup function handles det setup of the crypto part of the the VerkleTree with the elliptic curves and fields, takes as input a security parameter.
// It returns the public key.
func rootsetup(security, t int) pubParams {
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
	var params pubParams
	params.degree = t
	params.lagrangeBasis = make([]e.G1, t)
	params.diff2 = make([]e.G2, t)
	params.domain = make([]e.Scalar, t)
	params.aPrimeDomainI = make([]e.Scalar, t)
	for i := range params.lagrangeBasis {
		params.lagrangeBasis[i] = *e.G1Generator()
		params.diff2[i] = *e.G2Generator()
		params.domain[i] = *new(e.Scalar)
		params.aPrimeDomainI[i] = *new(e.Scalar)
	}

	params.zeroG1.SetUint64(0)
	params.oneG1.SetOne()

	ag2 := new(e.G2)
	ag2.ScalarMult(a, g2)
	//Generator from natural domain
	for i := range params.domain {
		params.domain[i].SetUint64(uint64(i))
	}

	for i := range params.aPrimeDomainI {
		params.aPrimeDomainI[i] = rootaPrime(params, i, params.aPrimeDomainI[i])
	}

	// lagrange Basis magic
	for i := range params.lagrangeBasis {
		// make evalLagrange func
		l := evalLagrangeValue(params, i, *a)
		params.lagrangeBasis[i].ScalarMult(&l, g1)
	}

	var e2 e.Scalar
	for i := range params.diff2 {
		e2.Sub(a, &params.domain[i])
		params.diff2[i].ScalarMult(&e2, g2)
	}

	//Creates the public key
	params = preCalculate(params)
	return params
}

func evalLagrangeValue(params pubParams, i int, v e.Scalar) e.Scalar {
	var ret e.Scalar
	ret.SetOne()
	numer := new(e.Scalar)
	denom := new(e.Scalar)
	elem := new(e.Scalar)
	for j := 0; j < params.degree; j++ {
		if j == i {
			continue
		}
		numer.Sub(&v, &params.domain[j])
		denom.Sub(&params.domain[i], &params.domain[j])
		denom.Inv(denom)
		elem.Mul(numer, denom)
		ret.Mul(&ret, elem)
	}
	return ret
}

func rootaPrime(params pubParams, m int, ret e.Scalar) e.Scalar {
	ret.SetOne()
	var eScaler e.Scalar

	for i := range params.domain {
		if i == m {
			continue
		}
		eScaler.Sub(&params.domain[m], &params.domain[i])
		ret.Mul(&ret, &eScaler)
	}
	return ret
}

func preCalculate(params pubParams) pubParams {
	params.precalc = &precompute{
		invsub: make([]e.Scalar, params.degree*2-1),
		ta:     make([][]e.Scalar, params.degree),
		tk:     make([]e.Scalar, params.degree),
	}

	for i := range params.precalc.ta {
		params.precalc.ta[i] = make([]e.Scalar, params.degree)
	}

	tj := new(e.Scalar)
	for j := 0; j < params.degree; j++ {
		tj.SetUint64(uint64(j))
		for m := 0; m < params.degree; m++ {
			if j == m {
				continue
			}
			idx := params.degree - 1 + m - j
			params.precalc.invsub[idx].SetUint64(uint64(m))
			params.precalc.invsub[idx].Sub(&params.precalc.invsub[idx], tj)
			params.precalc.invsub[idx].Inv(&params.precalc.invsub[idx])
		}
	}
	for j := range params.precalc.ta {
		for m := range params.precalc.ta[j] {
			if m == j {
				continue
			}
			params.precalc.ta[m][j].Set(&params.aPrimeDomainI[m])
			var invAprioriI e.Scalar
			invAprioriI.Inv(&params.aPrimeDomainI[j])
			params.precalc.ta[m][j].Mul(&params.precalc.ta[m][j], &invAprioriI)
			invSubmj := invSub(params, m, j)
			params.precalc.ta[m][j].Mul(&params.precalc.ta[m][j], &invSubmj) //TODO insert func

		}
	}

	for m := range params.precalc.tk {
		params.precalc.tk[m].SetUint64(0)
		for j := range params.precalc.ta[m] {
			if j == m {
				continue
			}
			params.precalc.tk[m].Add(&params.precalc.tk[m], &params.precalc.ta[m][j])
		}
	}
	return params
}

func invSub(params pubParams, m, j int) e.Scalar {
	//TODO maybe make nil case?
	idx := params.degree - 1 + m - j
	//TODO maybe make set if statement
	return params.precalc.invsub[idx]
}

func rootCommit(params pubParams, certs []e.Scalar) e.G1 {

	var elem e.G1
	ret := e.G1Generator()

	for i, e := range certs {
		elem.ScalarMult(&e, &params.lagrangeBasis[i])
		ret.Add(ret, &elem)
	}
	return *ret
}

func diff(params pubParams, vi, vj, ret e.Scalar) e.Scalar {
	//switch {
	//case vi.IsZero() == nil && vj == nil:
	//	ret.SetUint64(0)
	//	return ret //TODO maybe remove, as it probably does nothing.. We just don't dare.
	//case vi != nil && vj == nil:
	//	ret.Set(&vi)
	//case vi == nil && vj != nil:
	//	ret.Set(&vj)
	//	ret.Neg()
	//default:
	ret.Sub(&vi, &vj)
	//}
	return ret
}

func ta(params pubParams, m, j int, ret e.Scalar) e.Scalar {
	ret = invSub(params, m, j)
	ret.Mul(&ret, &params.aPrimeDomainI[m])
	var invNumb e.Scalar
	invNumb.Inv(&params.aPrimeDomainI[j])
	ret.Mul(&ret, &invNumb)
	return ret
}

func tk(params pubParams, m int, ret e.Scalar) e.Scalar {
	ret.SetUint64(0)
	var t e.Scalar
	for j := 0; j < params.degree; j++ {
		if j == m {
			continue
		}
		tempTA := ta(params, m, j, t)
		ret.Add(&ret, &tempTA)
	}
	return ret
}

func qPoly(params pubParams, certs []e.Scalar, i, m int, y, ret e.Scalar) e.Scalar {
	var numer e.Scalar
	if i != m {
		numer = diff(params, certs[m], y, numer) // TODO skriv number.sub(certs[m],y) da dette er legacy fra det andet kode
		if numer.IsEqual(&params.zeroG1) == 1 {
			ret.SetUint64(0)
			return ret
		}
		tempInvSub := invSub(params, m, i)
		ret.Mul(&numer, &tempInvSub)
		return ret
	}
	ret.SetUint64(0)
	var t e.Scalar
	var t1 e.Scalar
	for j := range certs {
		if j == m {
			continue
		}
		tempTA := ta(params, m, j, t1)
		t.Mul(&certs[j], &tempTA)
		ret.Add(&ret, &t)
	}
	tempTK := tk(params, m, t1)
	t.Mul(&certs[m], &tempTK)
	ret.Sub(&ret, &t)
	return ret

}

func rootProveGen(params pubParams, certs []e.Scalar, index int) e.G1 {

	ret := e.G1Generator()
	var o e.G1
	var qij e.Scalar
	for j := range params.domain {
		qij = qPoly(params, certs, index, j, certs[index], qij)
		o.ScalarMult(&qij, &params.lagrangeBasis[j])
		ret.Add(ret, &o)
	}
	return *ret
}

func rootVerify(params pubParams, commit, pi e.G1, v e.Scalar, index int) bool {
	p1 := e.Pair(&pi, &params.diff2[index])

	o := e.G1Generator()
	o.ScalarMult(&v, o)
	o.Neg()
	o.Add(&commit, o)
	g2Ident := e.G2Generator()
	p2 := e.Pair(o, g2Ident)
	fmt.Println("o:")
	fmt.Println(o)
	fmt.Println("ident:")
	fmt.Println(g2Ident)
	fmt.Println("p1: ", p1)
	fmt.Println("p2: ", p2)
	return p1.IsEqual(p2)

}

// Changes the certificates (bytes) to scalars and returns it as a list.
func rootcertToScalarVector(certs [][]byte) []e.Scalar {
	vects := make([]e.Scalar, len(certs))
	for i, v := range certs {
		vects[i].SetBytes(v)
	}
	return vects
}

// Evaluates the polynomium for a given x
func rootcalcPoly(x uint64, poly poly) e.Scalar {
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

// Calculates the divident used in buildtree. Takes the fanout and the unique combinations as input.
// Returns the dividents as a list.
// TODO redegør for det her med math!
func rootdividentCalculator(fanOut int, degreeComb [][][]int) []e.Scalar {
	dividentList := make([]e.Scalar, fanOut)
	var divident e.Scalar
	divident.SetOne()
	var iScalar e.Scalar
	for i := 0; i < fanOut; i++ {
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
	//TODO black magic
	for i := 1; i < fanOut; i++ {
		dividentMinusI.SetUint64(0)
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
			//dividentMinusI += sumDiv
			dividentMinusI.Add(&dividentMinusI, &sumDiv)
		}
		// inverse it so when we multiply with it, it will work as division!!!

		dividentMinusI.Inv(&dividentMinusI)

		dividentList[i] = dividentMinusI
	}
	return dividentList
}

func rootlagrangeBasisForGivenI(indexI int, fanOut int, dividentList []e.Scalar, degreeComb [][][]int, lagrangeBasisList *[][]e.Scalar) []e.Scalar {
	var coefToBe e.Scalar
	var combScalar e.Scalar
	dividentMinusI := dividentList[indexI]
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
				coefToBe.Mul(&coefToBe, &dividentMinusI)

				coefToBeList[j].Add(&coefToBe, &coefToBeList[j])
			}

		}
	}
	mutexBuddy.Lock()
	defer mutexBuddy.Unlock()
	(*lagrangeBasisList)[indexI] = coefToBeList
	return coefToBeList
}

func rootlagrangeBasisCalc(fanOut int, degreeComb [][][]int, dividentList []e.Scalar) [][]e.Scalar {
	// var lagrangeBasisList [][]e.Scalar
	lagrangeBasisList := make([][]e.Scalar, fanOut)
	var wg sync.WaitGroup
	for i := 0; i < fanOut; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lagrangeBasisForGivenI(i, fanOut, dividentList, degreeComb, &lagrangeBasisList)
		}()
		//numGoroutines := runtime.NumGoroutine()
		//fmt.Println("Number of active goroutines:", numGoroutines)

		//lagrangeBasisList = append(lagrangeBasisList, lagrangeBasisForGivenI(i, fanOut, dividentList, degreeComb, &lagrangeBasisList))
	}
	wg.Wait()
	//fmt.Println("LAGRANGEBASISLIST:", lagrangeBasisList)
	return lagrangeBasisList
}

// This translates the input vector into a polynomial which can be used for KZG commitment. It takes the scalar vector as input, unique combinations and dividentlist.
// It returns the polynomial of the vector, f(i)=scalarVect[i].
func rootrealVectorToPoly(scalarVect []e.Scalar, lagrangeBasisList [][]e.Scalar) poly {
	//Prepares variable for the polynomial.
	var answer poly
	coefs := make([]e.Scalar, len(scalarVect))
	coefs[0] = scalarVect[0] // first value in list of points, this is a constant coefficient in the polynomial (aka the first coefficient if a0 + a1x + a2x^2 + ...)
	var coefToBe e.Scalar
	//lagrangeBasisList := lagrangeBasisCalc(len(scalarVect), degreeComb, dividentList)
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
func rootquotientOfPoly(polynomial poly, x0 uint64) poly {
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
func rootcertVectorToPolynomial(certVect [][]byte, lagrangeBasisList [][]e.Scalar) poly {

	scalarVector := certToScalarVector(certVect)

	polynomial := realVectorToPoly(scalarVector, lagrangeBasisList)
	return polynomial
}

// Commit function, that computes the KZG polynomial commitment, given the public key and polynomial
func rootcommit(pk PK, polynomial poly) e.G1 {
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

func rootopen() int { //TODO fiks den aka. lav den
	return 0
}

// verifies that the polynomial for the commitment is correct, by recomputing the commitment for the polynomial and checking it is the same as the one to verify
func rootverifyPoly(pk PK, commitmentToVerify e.G1, polynomial poly) bool {
	commitment := commit(pk, polynomial)

	return commitment.IsEqual(&commitmentToVerify)
}

// Creates the witness for the specified index of the polynomial
// computing the quotientPolynomial for the given index and then calculating a commitment for the quotient and a evaluation of the index of the original polynomial
func rootcreateWitness(pk PK, polynomial poly, index uint64) witnessStruct {
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
func rootverifyWitness(pk PK, commitment e.G1, witness witnessStruct) bool {
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
