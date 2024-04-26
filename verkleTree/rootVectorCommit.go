package verkletree

import (
	"crypto/rand"

	//"fmt"
	//"runtime"
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
type rootWitnessStruct struct {
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
func rootSetup(security, t int) pubParams {
	//Sets up the generator elements, as well as the secret key a.
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	var params pubParams
	params.degree = t
	params.lagrangeBasis = make([]e.G1, t)
	params.diff2 = make([]e.G2, t)
	params.domain = make([]e.Scalar, t)
	params.aPrimeDomainI = make([]e.Scalar, t)
	for i := range params.lagrangeBasis {
		params.domain[i] = *new(e.Scalar)
		params.aPrimeDomainI[i] = *new(e.Scalar)
		params.lagrangeBasis[i] = *e.G1Generator()
		params.diff2[i] = *e.G2Generator()

	}

	params.zeroG1.SetUint64(0)
	params.oneG1.SetOne()

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
	var ret e.G1
	ret.SetIdentity()

	for i, e := range certs {
		elem.ScalarMult(&e, &params.lagrangeBasis[i])
		ret.Add(&ret, &elem)
	}
	return ret
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
	ret.Set(&params.precalc.ta[m][j])
	return ret
}

func tk(params pubParams, m int, ret e.Scalar) e.Scalar {
	ret.Set(&params.precalc.tk[m])
	return ret
	//ret.SetUint64(0)
	//var t e.Scalar
	//for j := 0; j < params.degree; j++ {
	//	if j == m {
	//		continue
	//	}
	//	tempTA := ta(params, m, j, t)
	//	ret.Add(&ret, &tempTA)
	//}
	//return ret
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

	var ret e.G1
	ret.SetIdentity()
	var o e.G1
	var qij e.Scalar
	for j := range params.domain {
		qij = qPoly(params, certs, index, j, certs[index], qij)
		o.ScalarMult(&qij, &params.lagrangeBasis[j])
		ret.Add(&ret, &o)
	}
	return ret
}

func rootVerify(params pubParams, commit, pi e.G1, v e.Scalar, index int) bool {
	p1 := e.Pair(&pi, &params.diff2[index])

	o := e.G1Generator()

	o.ScalarMult(&v, o)
	o.Neg()
	o.Add(&commit, o)
	g2Ident := e.G2Generator()
	p2 := e.Pair(o, g2Ident)
	return p1.IsEqual(p2)
}
