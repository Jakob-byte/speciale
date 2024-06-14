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

func vanderCalc(points []float64) [][]float64 {
	size := len(points)
	m := make([][]float64, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			m[i] = append(m[i], math.Pow(float64(i), float64(j)))
		}
	}
	return m
}

func calcDeterminant(vanderBoi [][]float64) float64 {
	det := 1
	for k := 1; k <= len(vanderBoi); k++ {
		for j := k + 1; j <= len(vanderBoi); j++ {
			det *= (j - k)
		}
	}
	return float64(det)
}

func newRealvVectorToPoly(points []float64) poly {
	var answer poly
	coefs := make([]float64, len(points))
	coefs[0] = points[0] // first value in list of points, this is constant coefficient
	divisor := 1.0
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
		divisor = newdivisorCalc(v, points, superX)
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
				//	fmt.Println("divisor: ", divisor)
				sumsum += (sumj * float64(sumk)) / (divisor)
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
			coefs[lambda] += (float64(sumsum) * points[v]) // divisor
			//	fmt.Println("new coefs: ", points[v], lambda, coefs[lambda])
		}
		if len(points)%2 == 1 {
			coefs[len(coefs)-1] += points[v] / divisor
		} else {
			coefs[len(coefs)-1] -= points[v] / divisor
		}
		//	fmt.Println("checkThis Out now: ", coefs[len(coefs)-1])
		//superCoefs[v] = coefs
	}
	//coefs[len(coefs)-1] = 1
	// -2.5+0.20
	//fmt.Println(answer.coefficients)
	//fmt.Println(divisor)
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

func newdivisorCalc(x int, points []float64, superX int) float64 {

	for i := 1; i <= x; i++ {
		superX = superX * 1 / (len(points) - i) * -(i) //Nice and fast.
	}
	return float64(superX)
}

func realVectorToPoly(points []float64) poly {
	var answer poly
	coefs := make([]float64, len(points))
	coefs[0] = points[0] // first value in list of points, this is constant coefficient
	divisor := 1.0

	//divisor Finder loop
	for i := range points {
		if i != 0 {
			divisor = divisor * float64(i)
		}
	}
	var degreeComb [][][]int
	for k := len(points) - 1; k > 0; k-- {
		degreeComb = append(degreeComb, combin.Combinations(len(points), k-1))
	}
	//flipBool := true
	var divisorMinusI float64
	var divToBe float64
	var sumDiv float64
	for i, y := range points {
		divisorMinusI = 0
		if i == 0 {
			divisorMinusI = divisor
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
				divisorMinusI += sumDiv
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
					//fmt.Println("Working coef: ", y, j+1, (coefToBe*y)/divisorMinusI)
					coefs[j+1] += (coefToBe * y) / divisorMinusI
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

func getCofactor(A [][]float64, temp [][]float64, p int, q int, n int) [][]float64 {

	i := 0
	j := 0

	// Looping for each element of the matrix
	for row := range n {

		for col := range n {

			// Copying into temporary matrix only those element
			// which are not in given row and column
			if row != p && col != q {

				temp[i][j] = float64(A[row][col])
				j += 1

				// Row is filled, so increase row index and
				// reset col index
				if j == n-1 {
					j = 0
					i += 1
				}
			}
		}
	}
	return temp
}

func determinant(A [][]float64, n int) float64 {

	D := 0.0 // Initialize result

	// Base case : if matrix contains single element
	if n == 1 {
		return A[0][0]
	}

	temp := make([][]float64, n) // To store cofactors
	for i := range n {
		temp[i] = make([]float64, n)
		// temp.append([None for _ in range(N)])
	}
	sign := 1.0 // To store sign multiplier

	// Iterate for each element of first row
	for f := range n {
		// Getting Cofactor of A[0][f]
		getCofactor(A, temp, 0, f, n)
		D += sign * A[0][f] * determinant(temp, n-1)

		// terms are to be added with alternate sign
		sign = -sign
	}

	return D
}

// Function to get adjoint of A[N][N] in adj[N][N].
func adjoint(A [][]float64, adj [][]float64, N int) [][]float64 {

	if N == 1 {
		adj[0][0] = 1
		return adj
	}

	// temp is used to store cofactors of A[][]
	sign := 1.
	temp := make([][]float64, N) // To store cofactors
	for i := range N {
		temp[i] = make([]float64, N)
		//temp.append([None for _ in range(N)])
	}

	for i := range N {

		for j := range N {
			// Get cofactor of A[i][j]
			getCofactor(A, temp, i, j, N)

			// sign of adj[j][i] positive if sum of row
			// and column indexes is even.
			//sign = [1, -1][(i + j) % 2]
			sign = 1
			if ((i + j) % 2) == 1 {
				sign = -1
			}

			// Interchanging rows and columns to get the
			// transpose of the cofactor matrix
			adj[j][i] = (sign) * (determinant(temp, N-1))
		}
	}
	return adj
}

func inverse(A [][]float64, inverse [][]float64, N int) [][]float64 {

	// Find determinant of A[][]
	det := calcDeterminant(A)
	//fmt.Println("calcdeter", det)
	//det := determinant(A, N)
	if det == 0 {
		print("Singular matrix, can't find its inverse")
		return [][]float64{}
	}

	// Find adjoint
	//adj := []
	adj := make([][]float64, N)
	for i := range N {
		adj[i] = make([]float64, N)
		//adj.append([None for _ in range(N)])
	}
	//fmt.Println("About to call adjoint")
	start := time.Now()
	adj = adjoint(A, adj, N)
	elapsed := time.Since(start)
	fmt.Println("ADJOINT RUNTIME FOR N", N, elapsed)
	// Find Inverse using formula "inverse(A) = adj(A)/det(A)"
	for i := range N {
		//fmt.Println("I am in inverse for I: ", i)
		for j := range N {
			inverse[i][j] = adj[i][j] / det
		}
	}
	return inverse
}

func MulMatrix(matrix1 [][]float64, matrix2 [][]float64) [][]float64 {
	result := make([][]float64, len(matrix1))
	for i := 0; i < len(matrix1); i++ {
		result[i] = make([]float64, len(matrix1))
		for j := 0; j < len(matrix2); j++ {
			for k := 0; k < len(matrix2); k++ {
				result[i][j] += matrix1[i][k] * matrix2[k][j]
			}
		}
	}
	return result
}
func MulMatrixMitVector(matrix1 [][]float64, vector []float64) []float64 {
	result := make([]float64, len(matrix1))
	for i := 0; i < len(matrix1); i++ {
		//result[i] = make([]float64, len(matrix1))
		for j := 0; j < len(vector); j++ {
			result[i] += matrix1[i][j] * vector[j]

		}
	}
	return result
}

func main() {
	points := []float64{
		5,
		15,
		//	9,
		//	29,
		//	45,
		//	23,
		//	51,
		//	52,
		//	69,
		//	190,
		//	16,
		//	52,
	}

	//poly2 := realVectorToPoly(points)
	//quotientPoly := quotientOfPoly(poly2, 2)
	//fmt.Println("coefffs", poly2.coefficients)
	fmt.Println("NEW Vander-WAY COMING NOW")
	vanderboii := vanderCalc(points)
	for i := range len(vanderboii) {
		fmt.Println(vanderboii[i])
	}

	deter := calcDeterminant(vanderboii)
	fmt.Println("deterboiiiii", deter)
	//quotientPoly := quotientOfPoly(poly2, 2)
	fmt.Println("Succes")

	fmt.Println("TRYING NEW DETERMINANT FUNC NOW")
	N := len(vanderboii)
	//det := determinant(vanderboii, N)
	//fmt.Println("DETDETDET: ", det)
	adj := make([][]float64, N) // To store cofactors
	inv := make([][]float64, N)
	for i := range N {
		adj[i] = make([]float64, N)
		inv[i] = make([]float64, N)
		//temp.append([None for _ in range(N)])
	}
	//adjencyCalc := adjoint(vanderboii, adj, N)
	//fmt.Println("ADJ: ", adjencyCalc)
	fmt.Println("Calling inverse now")
	start := time.Now()
	invCalc := inverse(vanderboii, inv, N)
	timed := time.Since(start)
	fmt.Println("timed", timed)
	fmt.Println("INVERSECALC: ", invCalc)
	//shouldBeIdentity := MulMatrix(vanderboii, invCalc)
	//for i := range N {
	//	fmt.Println("ooohhh", shouldBeIdentity[i][i])
	//}
	shouldBeFunc := MulMatrixMitVector(invCalc, points)
	fmt.Println("should be the coeff", shouldBeFunc)
	for i := 1; i <= 40; i++ {
		points = append(points, float64(i))
		vanderboii := vanderCalc(points)

		N := len(vanderboii)
		//det := determinant(vanderboii, N)
		//fmt.Println("DETDETDET: ", det)
		adj := make([][]float64, N) // To store cofactors
		inv := make([][]float64, N)
		for j := range N {
			adj[j] = make([]float64, N)
			inv[j] = make([]float64, N)
			//temp.append([None for _ in range(N)])
		}
		start := time.Now()
		inverse(vanderboii, inv, N)
		elapsed := time.Since(start)
		fmt.Println("Time spent on fanout", len(points), " is ", elapsed)
	}

}
