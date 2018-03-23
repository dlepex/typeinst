// nolint
package maps

import (
	"fmt"
)

type Node struct {
	key   K
	val   V
	h     int
	next  **Node
	fakes [42]wrapper
}

type wrapper struct {
	*Node
}

type Maps struct{}

func (_ Maps) Create(keys []K, mapper func(K) V) map[K]V {
	return nil
}

type Maps2 struct{}

func (_ Maps2) Create2(keys []K, mapper func(K) V) map[K]V {
	x := maxW
	fmt.Print(x)
	switch len(keys) {
	case maxW:
		return nil
	}
	return make(map[K]V)
}
