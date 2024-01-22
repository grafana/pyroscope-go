package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/grafana/pyroscope-go"
)

//go:noinline
func isPrime(n int64) bool {
	for i := int64(2); i <= n; i++ {
		if i*i > n {
			return true
		}
		if n%i == 0 {
			return false
		}
	}
	return false
}

//go:noinline
func slow(n int64) int64 {
	sum := int64(0)
	for i := int64(1); i <= n; i++ {
		sum += i
	}
	return sum
}

//go:noinline
func fast(n int64) int64 {
	sum := int64(0)
	root := int64(math.Sqrt(float64(n)))
	for a := int64(1); a <= n; a += root {
		b := a + root - 1
		if n < b {
			b = n
		}
		sum += (b - a + 1) * (a + b) / 2
	}
	return sum
}

//go:noinline
func slow0(n int64) int64 { return slow(n) }

//go:noinline
func slow1(n int64) int64 { return slow(n) }

//go:noinline
func slow2(n int64) int64 { return slow(n) }

//go:noinline
func slow3(n int64) int64 { return slow(n) }

//go:noinline
func slow4(n int64) int64 { return slow(n) }

//go:noinline
func slow5(n int64) int64 { return slow(n) }

//go:noinline
func slow6(n int64) int64 { return slow(n) }

//go:noinline
func slow7(n int64) int64 { return slow(n) }

//go:noinline
func slow8(n int64) int64 { return slow(n) }

//go:noinline
func slow9(n int64) int64 { return slow(n) }

//go:noinline
func slow10(n int64) int64 { return slow(n) }

//go:noinline
func slow11(n int64) int64 { return slow(n) }

//go:noinline
func slow12(n int64) int64 { return slow(n) }

//go:noinline
func slow13(n int64) int64 { return slow(n) }

//go:noinline
func slow14(n int64) int64 { return slow(n) }

//go:noinline
func slow15(n int64) int64 { return slow(n) }

//go:noinline
func fast0(n int64) int64 { return fast(n) }

//go:noinline
func fast1(n int64) int64 { return fast(n) }

//go:noinline
func fast2(n int64) int64 { return fast(n) }

//go:noinline
func fast3(n int64) int64 { return fast(n) }

//go:noinline
func fast4(n int64) int64 { return fast(n) }

//go:noinline
func fast5(n int64) int64 { return fast(n) }

//go:noinline
func fast6(n int64) int64 { return fast(n) }

//go:noinline
func fast7(n int64) int64 { return fast(n) }

//go:noinline
func fast8(n int64) int64 { return fast(n) }

//go:noinline
func fast9(n int64) int64 { return fast(n) }

//go:noinline
func fast10(n int64) int64 { return fast(n) }

//go:noinline
func fast11(n int64) int64 { return fast(n) }

//go:noinline
func fast12(n int64) int64 { return fast(n) }

//go:noinline
func fast13(n int64) int64 { return fast(n) }

//go:noinline
func fast14(n int64) int64 { return fast(n) }

//go:noinline
func fast15(n int64) int64 { return fast(n) }

//go:noinline
func isPrime0(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime1(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime2(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime3(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime4(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime5(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime6(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime7(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime8(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime9(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime10(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime11(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime12(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime13(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime14(n int64) bool { return isPrime(n) }

//go:noinline
func isPrime15(n int64) bool { return isPrime(n) }

//go:noinline
func slowMux(n, d int64) int64 {
	switch d {
	case 0:
		return slow0(n)
	case 1:
		return slow1(n)
	case 2:
		return slow2(n)
	case 3:
		return slow3(n)
	case 4:
		return slow4(n)
	case 5:
		return slow5(n)
	case 6:
		return slow6(n)
	case 7:
		return slow7(n)
	case 8:
		return slow8(n)
	case 9:
		return slow9(n)
	case 10:
		return slow10(n)
	case 11:
		return slow11(n)
	case 12:
		return slow12(n)
	case 13:
		return slow13(n)
	case 14:
		return slow14(n)
	default:
		return slow15(n)
	}
}

//go:noinline
func fastMux(n, d int64) int64 {
	switch d {
	case 0:
		return fast0(n)
	case 1:
		return fast1(n)
	case 2:
		return fast2(n)
	case 3:
		return fast3(n)
	case 4:
		return fast4(n)
	case 5:
		return fast5(n)
	case 6:
		return fast6(n)
	case 7:
		return fast7(n)
	case 8:
		return fast8(n)
	case 9:
		return fast9(n)
	case 10:
		return fast10(n)
	case 11:
		return fast11(n)
	case 12:
		return fast12(n)
	case 13:
		return fast13(n)
	case 14:
		return fast14(n)
	default:
		return fast15(n)
	}
}

//go:noinline
func isPrimeMux(n, d int64) bool {
	switch d {
	case 0:
		return isPrime0(n)
	case 1:
		return isPrime1(n)
	case 2:
		return isPrime2(n)
	case 3:
		return isPrime3(n)
	case 4:
		return isPrime4(n)
	case 5:
		return isPrime5(n)
	case 6:
		return isPrime6(n)
	case 7:
		return isPrime7(n)
	case 8:
		return isPrime8(n)
	case 9:
		return isPrime9(n)
	case 10:
		return isPrime10(n)
	case 11:
		return isPrime11(n)
	case 12:
		return isPrime12(n)
	case 13:
		return isPrime13(n)
	case 14:
		return isPrime14(n)
	default:
		return isPrime15(n)
	}
}

//go:noinline
func main() {
	pyroscopeProfiler, _ := pyroscope.Start(pyroscope.Config{
		ApplicationName: "timing-demo",
		ServerAddress:   "http://localhost:4040",
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
		},
		Logger: pyroscope.StandardLogger,
	})
	fmt.Println(pyroscopeProfiler)
	// defer pyroscopeProfiler.Stop()

	startTime := time.Now()
	base := rand.Int63n(1000000) + 1
	for i := int64(0); i < 40000000; i++ {
		secs := int64(time.Since(startTime) / time.Second)
		if secs > 15 {
			break
		}
		n := rand.Int63n(10000) + 1

		if isPrimeMux(base+i, secs) {
			slowMux(n, secs)
		} else {
			fastMux(n, secs)
		}
	}
}
