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

	for i := len(points) - 1; i > 1; i-- {
		superX = superX * i
	}

	for v := 0; v < len(points); v++ {
		fmt.Println("CALCULATING DIVIDENT")
		divident = newDividentCalc(v, points, superX)

		fmt.Println("DONE DIVIDENT", divident)
		for lambda := 1; lambda < len(points); lambda++ {

			//fmt.Println(coefs[202020])
			weBroke := false
			sumsum := 0.0
			for i := 1; i < len(points)-lambda; i++ {
				fmt.Println("In i loop")
				sumj := 1.
				for j := i; j < i+lambda; j++ {
					fmt.Println("In j loop")
					if v == j {
						//if v >= j && v <= j + lambda
						weBroke = true
						break
					}
					sumj *= float64(j)
				}
				if weBroke {
					weBroke = false
					continue
				}
				sumk := 0.0
				for k := i + lambda; k < len(points); k++ {
					fmt.Println("In k loop")
					if k == v { //Måske yeet, virker unødvendigt.
						continue
					}
					sumk += float64(k)
				}
				sumsum = sumj * sumk
			}
			if (lambda % 2) == 0 {
				coefs[lambda] -= sumsum / divident
			} else {
				coefs[lambda] += sumsum / divident
			}
		}
	}
	answer.coefficients = coefs
	fmt.Println(answer.coefficients)
	fmt.Println(divident)

	return answer
}

func newDividentCalc(x int, points []float64, superX int) float64 {
	sumsumsum := 0
	for lambda := 2; lambda < len(points); lambda++ {
		sumsumK := 0

		for i := 1; i < (len(points) - lambda + 1); i++ {
			xLambdaV := math.Pow(float64(x), float64(lambda))
			fmt.Println("lambda boooooy", xLambdaV)
			sumj := 1
			for j := i + 1; j < (i + len(points) - lambda - 1); j++ {
				sumj *= j
			}
			sumj = sumj * int(xLambdaV)
			sumK := 0
			for k := (len(points) - lambda + i - 1); k < len(points); k++ { //TODO This doesn't work.. at all..
				if x == k {
					continue
				}
				fmt.Println("iamK", k)
				sumK += k
			}
			fmt.Println("SUMJ", sumj)
			fmt.Println("SUMK", sumK)

			sumsumK += sumj * sumK

			fmt.Println("SUMsumK", sumsumK)
		}

		if (lambda % 2) == 0 {
			sumsumK *= -1
		}
		fmt.Println("SUMsumK", sumsumK)

		sumsumsum += sumsumK
	}
	fmt.Println("sumsumsumsum", sumsumsum)

	sumsumsum += superX
	return float64(sumsumsum)
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
					fmt.Println("I AM DIVIDENT FOR j+1", j+1, dividentMinusI)
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
		20,
	}

	poly2 := realVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	fmt.Println("coefffs", poly2.coefficients)
	fmt.Println("NEW WAY COMING NOW")
	poly3 := newRealvVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	fmt.Println("coefffs", poly3.coefficients)
	fmt.Println("Succes")

}
