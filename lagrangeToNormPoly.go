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

type lagrangePoly struct {
	coefficients []float64
}

func vectorToPoly(points []float64) lagrangePoly {
	var answer lagrangePoly
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

func realVectorToPoly(points []float64) lagrangePoly {
	var answer lagrangePoly
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
						for _, blabla := range comb {
							divToBe *= float64(blabla)
						}
						if len(comb) == 0 {
							divToBe *= math.Pow(float64(i), float64(j+1))

						} else {
							divToBe *= math.Pow(float64(i), float64(j+1))
						}
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
					for _, blabla := range comb {
						coefToBe *= float64(blabla)
					}

					if ((j) % 2) == 0 {
						coefToBe *= -1
					}

					coefs[j+1] += (coefToBe * y) / dividentMinusI
				}

			}

			coefs[j+1] = (coefs[j+1])

		}

		//c1 := y*9
	}
	fmt.Println("coefs", coefs)
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

func calcPoly(x float64, poly lagrangePoly) float64 {
	var answer float64
	for i, a := range poly.coefficients {
		answer = answer + a*math.Pow(x, float64(i))
	}
	return answer
}





func main() {
	points := []float64{
		5,
		150,
		9,
		21,
		170000,
		59,
		10,
	}
	fmt.Println(points)
	poly := vectorToPoly(points)
	fmt.Println("coefficients", poly.coefficients)
	fmt.Println("results x=1", calcPoly(0, poly))
	fmt.Println("results x=2", calcPoly(1, poly))
	fmt.Println("results x=3", calcPoly(2, poly))
	poly2 := realVectorToPoly(points)
	fmt.Println(poly2.coefficients)
	fmt.Println("results x=0", calcPoly(0, poly2))
	fmt.Println("results x=1", calcPoly(1, poly2))
	fmt.Println("results x=2", calcPoly(2, poly2))
	fmt.Println("results x=3", calcPoly(3, poly2))
	fmt.Println("results x=4", calcPoly(4, poly2))
	fmt.Println("results x=5", calcPoly(5, poly2))
	fmt.Println("results x=6", calcPoly(6, poly2))

	fmt.Println("Succes")

}

5 + 63644*x+ 1.40008+06*x^2 .... -f(x_0)
/
x - x_0


