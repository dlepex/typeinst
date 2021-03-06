// nolint
package maps

import (
	"fmt"
)

type K = interface{}
type V = interface{}

type Map map[K]V

func (m Map) KeyValues(keys *[]K, values *[]V) {
	for k, v := range m {
		if keys != nil {
			*keys = append(*keys, k)
		}
		if values != nil {
			*values = append(*values, v)
		}
	}
}

const helloWorld uint32 = 20
const helloWorld1 string = "a"

type TreeMap struct {
	l   *Node
	r   *Node
	len int
}

func newTreeMap() *TreeMap {
	return nil
}

func create(keys []K, values []V) *TreeMap {
	return nil
}

func (t *TreeMap) Put(k K, v V) {
	fmt.Printf("%v %s", helloWorld, helloWorld1)
}

const maxW = 99

func makeWrappers() []*wrapper {
	return nil
}

func makeWrappers99() []**[maxW]*[]wrapper {
	return nil
}
