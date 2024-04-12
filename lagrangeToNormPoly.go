package main

import (
	"fmt"
	"math"
	"slices"
	"time"

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

	//superCoefs := make([][]float64, len(points))

	for i := len(points) - 1; i > 1; i-- {
		superX = superX * i
	}

	for v := 0; v < len(points); v++ {
		//fmt.Println("AT point v: ", v)
		//if v == 4 {
		//	fmt.Println(coefs[100])
		//}
		//if v > 0 {
		//	coefs = make([]float64, len(points))
		//	coefs[0] = 0
		//}
		//fmt.Println("v: ", v)
		divident = newDividentCalc(v, points, superX)
		sumsum := 0.0
		for lambda := 1; lambda < len(points)-1; lambda++ {

			//fmt.Println("At lambda: ", lambda)
			sumsum = 0
			updatedJList := true
			//fmt.Println(updatedJList)

			jList := make([]float64, len(points)-2-(lambda))
			//fmt.Println("lenOFJlistHERE: ", len(jList))
			ctrJlist := 0.
			for o := range jList { //init the jList to be 1,2,3,4,....
				if o == v {
					ctrJlist++
				}
				jList[o] = float64(o + int(ctrJlist))
			}

			if len(jList) != 0 && v != 0 {
				appendNumb := jList[len(jList)-1] + 1
				if int(appendNumb) == v {
					appendNumb++
				}
				jList = append(jList[1:], appendNumb)

			}
			i := 1
			for updatedJList {
				//		fmt.Println("jList:", jList)
				//if v > 0 && i == 1 {
				//	ctr += 1
				//
				//	continue
				//}

				sumj := 1.
				for _, j := range jList { //j := i; j < len(points)-lambda-1+ctr; j++ {
					//if v == j {
					//	continue
					//}
					sumj *= j
				}
				sumk := 0
				//TODO start at end of jList for k Loop

				if len(jList) == 0 {
					jList = append(jList, 0)
				}
				for k := int(jList[len(jList)-1]) + 1; k < len(points); k++ { //len(points) - lambda - 1 + ctr; k < len(points); k++ { // kommer til at plusse sig selv til i nogle cases! KlemZe har added et if statement
					if v == k {
						continue
					}
					sumk += k
				}

				//	fmt.Println("lambda: ", lambda)
				//	fmt.Println("sumj: ", sumj)
				//	fmt.Println("sumk: ", sumk)
				//	fmt.Println("divident: ", divident)
				sumsum += (sumj * float64(sumk)) / (divident)
				//	fmt.Println("sumsum: ", sumsum)
				if lambda == len(points)-2 {
					//			fmt.Println("Vi er her, dudes!")
					break
				}

				//2-1 ; o >= 0 ; o--
				//
				for o := len(jList) - 1; o >= 0; o-- { //update jList for next subexpression e.g. (1,2,3) -> (1,2,4)
					//		fmt.Println("funky math", float64(len(points))-float64(2+(len(jList)-1-o)))
					funkymath := float64(len(points)) - float64(2+(len(jList)-1-o))
					//		fmt.Println("o:", o)
					if jList[o] < funkymath { // 6-(2+(3-1-2) =4 ---- 6-(2+(3-1-1)) ---6 -(2+(3-1-0))
						// jList[1]=3 < 5-(2+(2-1-1))=3  ----- o=1
						// jList[0]=2 < 5-(2+(2-1-0))=2  ----- o=0
						if jList[o]+1 == float64(v) {
							jList[o] += 1

						}
						jList[o]++
						if jList[o] > funkymath && o == 0 {
							updatedJList = false
							break
						}
						if o != len(jList)-1 {
							ctr42 := 0
							for q := o + 1; q < len(jList); q++ { // <=?
								ctr42 = 0
								if jList[q-1]+1 == float64(v) {
									ctr42++ //no work
								}
								jList[q] = jList[q-1] + 1 + float64(ctr42)

							}
						}
						//if jList[len(jList)-1] > funkymath+1 {
						//	updatedJList = false
						//}
						break
					}
					//		fmt.Println("WE SET TO FALSE", i, o, jList)

					if o == 0 {
						updatedJList = false
					}
				}
				i++
			}
			if lambda%2 == 1 {
				sumsum *= -1
			}
			coefs[lambda] += (float64(sumsum) * points[v]) // divident
			//	fmt.Println("new coefs: ", points[v], lambda, coefs[lambda])
		}
		if len(points)%2 == 1 {
			coefs[len(coefs)-1] += points[v] / divident
		} else {
			coefs[len(coefs)-1] -= points[v] / divident
		}
		//	fmt.Println("checkThis Out now: ", coefs[len(coefs)-1])
		//superCoefs[v] = coefs
	}
	//coefs[len(coefs)-1] = 1
	// -2.5+0.20
	//fmt.Println(answer.coefficients)
	//fmt.Println(divident)
	//finalCoef := make([]float64, len(points))
	//finalCoef[0] = points[0]
	//fmt.Println("superCoefs: ", superCoefs)
	//for i, v := range points {
	//	for j, u := range superCoefs[i] {
	//		fmt.Println("u:", u)
	//		fmt.Println("v:", v)
	//		uv := u * v
	//		fmt.Println("uv:", v)
	//		finalCoef[j] += uv
	//	}
	//}
	//
	//answer.coefficients = finalCoef
	answer.coefficients = coefs
	return answer
}

func newDividentCalc(x int, points []float64, superX int) float64 {

	for i := 1; i <= x; i++ {
		superX = superX * 1 / (len(points) - i) * -(i) //Nice and fast.
	}
	return float64(superX)
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
					//fmt.Println("Working coef: ", y, j+1, (coefToBe*y)/dividentMinusI)
					coefs[j+1] += (coefToBe * y) / dividentMinusI
					//fmt.Println("coef[j+1]: ", coefs[j+1])
				}
				//
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
		29,
	}

	//poly2 := realVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	//fmt.Println("coefffs", poly2.coefficients)
	fmt.Println("NEW WAY COMING NOW")
	poly3 := newRealvVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	fmt.Println("coefffs", poly3.coefficients)
	fmt.Println("Succes")

	for i := 1; i <= 40; i++ {
		points = append(points, float64(i))
		start := time.Now()
		newRealvVectorToPoly(points)
		timeSpent := time.Since(start)
		fmt.Println("Time spent on fanout", len(points), " is ", timeSpent)
	}
	//fmt.Println("realCoefs: ", poly2.coefficients)

}
