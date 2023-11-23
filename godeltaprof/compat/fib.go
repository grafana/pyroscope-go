//go:build go1.18

package compat

type num interface{ int | int64 }

func fib[N num](n *N) *N {
	v := *n
	if v < 2 {
		return &v
	}
	v1 := v - 1
	n1 := fib(&v1)
	v2 := v - 2
	n2 := fib(&v2)
	res := *n1 + *n2
	return &res
}
