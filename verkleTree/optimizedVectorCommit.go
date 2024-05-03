package verkletree

import (
	"crypto/rand"

	//"fmt"
	//"runtime"

	e "github.com/cloudflare/circl/ecc/bls12381"
)

// The witness struct which contains the necessary info for a witness to prove it is contained in a commitment.
type optimizedWitnessStruct struct {
	Index uint64
	Fx0   e.Scalar
	W     e.G1
}

// A struct containing the precomputed values used for computing the commitment
// TODO maybe rewrite after knowing what these values actually are
type precompute struct {
	invsub []e.Scalar
	ta     [][]e.Scalar
	tk     []e.Scalar
}

// the public key/public paramters containing the nessasary values for calculating the vectorcommit/prove
// TODO ??? hvad er de her values faktisk, vi har bare skrevet dem ind
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

// type 3 kzg setting https://www.zkdocs.com/docs/zkdocs/commitments/kzg_polynomial_commitment/
// setup from https://hackmd.io/@Evaldas/SJ9KHoDJF and https://github.com/lunfardo314/verkle
// The setup function handles det setup of the crypto part of the the VerkleTree with the elliptic curves and fields, takes as input a security parameter.
// It returns the public paramters.
func optimizedSetup(security, t int) pubParams {
	//Sets up the generator elements, as well as the secret key a.
	g1 := e.G1Generator()
	g2 := e.G2Generator()
	a := new(e.Scalar) //secretkey a, is forgotten and destroyed.
	a.Random(rand.Reader)

	// setups and defines the slices needed in the public parameters
	var params pubParams
	params.degree = t
	params.lagrangeBasis = make([]e.G1, t)
	params.diff2 = make([]e.G2, t)
	params.domain = make([]e.Scalar, t)
	params.aPrimeDomainI = make([]e.Scalar, t)
	for i := range params.lagrangeBasis {
		params.domain[i] = *new(e.Scalar)
		params.aPrimeDomainI[i] = *new(e.Scalar)
		params.lagrangeBasis[i] = *g1
		params.diff2[i] = *g2 // TODO må vi sætte dem her til g1 og g2 variablerne vi har lavet tidligere er ødelægger det noget med pointer magi?
	}

	params.zeroG1.SetUint64(0)
	params.oneG1.SetOne()

	//Generator from natural domain
	for i := range params.domain {
		params.domain[i].SetUint64(uint64(i))
	}

	// TODO Generator from aPrimeDomain  what is this domain?
	for i := range params.aPrimeDomainI {
		params.aPrimeDomainI[i] = optimizedAPrime(params, i, params.aPrimeDomainI[i])
	}

	// lagrange Basis magic
	// evaluate the lagrange basis for the given value up to the degree of the polynomial(size of vector to commit)
	for i := range params.lagrangeBasis {
		l := evalLagrangeValue(params, i, *a)
		params.lagrangeBasis[i].ScalarMult(&l, g1)
	}

	//TODO more precompation what is this?
	var e2 e.Scalar
	for i := range params.diff2 {
		e2.Sub(a, &params.domain[i])
		params.diff2[i].ScalarMult(&e2, g2)
	}

	// TODO calls the preCalculate function that calculates the invsub,ta & tk
	params = preCalculate(params)

	//Returns the public key/paramters
	return params
}

// Function to evaluate the lagrange value in the secret value for the given index and secret key a
func evalLagrangeValue(params pubParams, i int, a e.Scalar) e.Scalar {
	var ret e.Scalar
	ret.SetOne()
	numer := new(e.Scalar)
	denom := new(e.Scalar)
	elem := new(e.Scalar)
	for j := 0; j < params.degree; j++ {
		if j == i {
			continue
		}
		// TODO does some magic
		numer.Sub(&a, &params.domain[j])
		denom.Sub(&params.domain[i], &params.domain[j])
		denom.Inv(denom)
		elem.Mul(numer, denom)
		ret.Mul(&ret, elem)
	}
	return ret
}

// TODO function to do some magic? in the public parameters
func optimizedAPrime(params pubParams, m int, ret e.Scalar) e.Scalar {
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

// TODO precalculates values used to find the quotient polynomial?
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

// TODO what does this do?
func invSub(params pubParams, m, j int) e.Scalar {
	idx := params.degree - 1 + m - j
	return params.precalc.invsub[idx]
}

// Calculates the commitment given the pubParams for a given a slice of certs with type e.Scalar
func optimizedCommit(params pubParams, certs []e.Scalar) e.G1 {

	var elem e.G1
	var ret e.G1
	ret.SetIdentity()

	for i, e := range certs {
		elem.ScalarMult(&e, &params.lagrangeBasis[i])
		ret.Add(&ret, &elem)
	}
	return ret
}

func ta(params pubParams, m, j int, ret e.Scalar) e.Scalar {
	ret.Set(&params.precalc.ta[m][j])
	return ret
}

func tk(params pubParams, m int, ret e.Scalar) e.Scalar {
	ret.Set(&params.precalc.tk[m])
	return ret
}

// TODO calculates the quotient polynomial, used for calculating the proof
func qPoly(params pubParams, certs []e.Scalar, i, m int, y, ret e.Scalar) e.Scalar {
	var numer e.Scalar
	if i != m {
		//numer = diff(certs[m], y, numer) // TODO skriv number.sub(certs[m],y) da dette er legacy fra det andet kode
		numer.Sub(&certs[m], &y)
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

// generates the proof for a index in the given vector of e.Scalar values
func optimizedProveGen(params pubParams, certs []e.Scalar, index int) e.G1 {

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

// Verifies the commitment and proof
// TODO refactored Pi to Proof, is that OK ? code makes more sense now
// TODO refactored v to vi as it is the vector at index i
func optimizedVerify(params pubParams, commit, proof e.G1, vi e.Scalar, index int) bool {
	p1 := e.Pair(&proof, &params.diff2[index])

	o := e.G1Generator()

	o.ScalarMult(&vi, o)
	o.Neg()
	o.Add(&commit, o)
	g2Ident := e.G2Generator()
	p2 := e.Pair(o, g2Ident)
	return p1.IsEqual(p2)
}
