package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
)

func main() {

	p, _ := rand.Prime(rand.Reader, 64)
	q, _ := rand.Prime(rand.Reader, 64)

	e := new(big.Int).SetInt64(65537)
	Mval := "10"

	argCount := len(os.Args[1:])

	if argCount > 0 {
		Mval = os.Args[1]
	}

	M, _ := new(big.Int).SetString(Mval, 10) // Base 10

	N := new(big.Int).Mul(p, q)

	C := new(big.Int).Exp(M, e, N)

	Pminus1 := new(big.Int).Sub(p, new(big.Int).SetInt64(1))
	Qminus1 := new(big.Int).Sub(q, new(big.Int).SetInt64(1))
	PHI := new(big.Int).Mul(Pminus1, Qminus1)

	d := new(big.Int).ModInverse(e, PHI)

	Plain := new(big.Int).Exp(C, d, N)

	fmt.Printf("M= %s\n", M)
	fmt.Printf("p= %s\n", p)
	fmt.Printf("q= %s\n", q)

	fmt.Printf("N= %s\n", N)
	fmt.Printf("Public: e= %s\n", e)
	fmt.Printf("Private: d= %s\n", d)
	fmt.Printf("\nCipher C= %s\n", C)
	fmt.Printf("\nDecrypt Plain= %s\n", Plain)

}
