package main

import (
	"fmt"
	"math"
	"slices"

	combin "gonum.org/v1/gonum/stat/combin"
)

type point struct {
	y float64
}

type poly struct {
	coefficients []float64
}

func vectorToPoly(points []float64) poly {
	var answer poly
	//now we do magic!!! :)

	b1 := (points[0] * 1 * 2) / (1 * 2)
	b2 := (points[0] * 2) / (1 * 2)
	b3 := points[0] / (1 * 2)
	b4 := points[0] / (1 * 2)
	v0Coefs := []float64{b1, -b2 - b3, b4}
	c1 := (points[1] * 0 * 2) / (-1*2 + 1)
	c2 := (points[1] * 2) / (-1*2 + 1)
	c3 := (points[1] * 0) / (-1*2 + 1)
	c4 := points[1] / (-1*2 + 1)
	v1Coefs := []float64{c1, -c2 - c3, c4}
	d1 := (points[2] * 0 * 1) / (-1*2 + 4)
	d2 := (points[2] * 1) / (-1*2 + 4)
	d3 := (points[2] * 0) / (-1*2 + 4)
	d4 := points[2] / (-1*2 + 4)
	v2Coefs := []float64{d1, -d2 - d3, d4}

	listlist := make([]float64, len(points))
	for i := 0; i < len(points); i++ {
		if i < 3 {
			listlist[i] = v0Coefs[i] + v1Coefs[i] + v2Coefs[i]
		}
	}
	answer.coefficients = listlist
	return answer
}

func newRealvVectorToPoly(points []float64) poly {
	var answer poly
	coefs := make([]float64, len(points))
	coefs[0] = points[0] // first value in list of points, this is constant coefficient
	divident := 1.0
	superX := 1

	for i := len(points); i > 1; i-- {
		superX = superX * i
	}

	for i := 0; i < len(points); i++ {
		for coefIndex := 1; coefIndex < len(points); coefIndex++ {
			//set of length coefindex
			// for loop which adds combinations to the set?
			// 2 extra for loops?
			for k := 1; k < len(points); k++ {
				coefs[coefIndex] += (float64(superX) * (1 / float64(k))) / float64(divident)
			}

			if coefIndex%2 == 1 {
				coefs[coefIndex] *= -1
			}
		}
	}

	fmt.Println(answer)
	fmt.Println(divident)

	return answer
}

func realVectorToPoly(points []float64) poly {
	var answer poly
	coefs := make([]float64, len(points))
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
		for j, combs := range degreeComb {
			for _, comb := range combs {
				if !slices.Contains(comb, i) {
					coefToBe := 1.0
					for _, c := range comb {
						coefToBe *= float64(c)
					}

					if ((j) % 2) == 0 {
						coefToBe *= -1
					}
					coefs[j+1] += (coefToBe * y) / dividentMinusI
					//fmt.Println("coef[j+1]: ", coefs[j+1])
				}

			}

		}

		//c1 := y*9
	}
	//fmt.Println("coefs", coefs)
	answer.coefficients = coefs
	return answer
}

//func recursionYay(currentVal float64, stepsLeft int, amountToMult int, myNumb int) float64 {
//	if stepsLeft > 0 {
//
//		// x2 *x3 + x1*x3 + x1*x2
//		currentVal = currentVal * float64(stepsLeft)
//		return recursionYay(currentVal, stepsLeft-1)
//	} else {
//		return currentVal
//	}
//}

func calcPoly(x float64, poly poly) float64 {
	var answer float64
	for i, a := range poly.coefficients {
		answer = answer + a*math.Pow(x, float64(i))
		//fmt.Println("Answer in I ", i, answer)
	}
	return answer
}

func quotientOfPoly(polynomial poly, x0 float64) poly {
	var quotient poly
	degree := len(polynomial.coefficients)
	coefs := make([]float64, degree-1)
	//fmt.Println("coefficients len: ", len(polynomial.coefficients))
	for i, _ := range polynomial.coefficients[1:] { //we ignore the forst coeff as it is divided out
		//fmt.Println("i:", i)
		count := 0
		for j := i; j < len(coefs); j++ {
			//fmt.Println("j:", j)
			coefs[i] += polynomial.coefficients[j+1] * math.Pow(x0, float64(count))
			count++
			//fmt.Println("OG coefs: ", polynomial.coefficients[j+1])
			//fmt.Println("coefs[i]=v", i, coefs[i])
		}
	}
	quotient.coefficients = coefs
	return quotient
}

func main() {
	points := []float64{
		5,
		15,
		9,
		//27,
	}

	poly2 := realVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	fmt.Println("coefffs", poly2.coefficients)
	fmt.Println("Succes")

}
