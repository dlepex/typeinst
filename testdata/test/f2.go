package test

import (
	"fmt"
	_un "unicode"
)

type T struct{} //typeinst: typevar

type E = interface{} //typeinst: typevar

type Slice []T

type methoder struct {
	s string
}
type Node struct {
	el E
	x  interface{}
}

func NewNode(t T, n Node) Node {
	_un.IsNumber('0')
	_, ok := n.x.(Slice)
	fmt.Println(ok)
	return Node{}
}
func NodeToSlice(n Node) *Slice {
	return nil
}

func (a Slice) IndexOf(el T) int {
	_un.IsNumber('0')
	for i, v := range a {
		if v == el {
			return i
		}
	}
	return -1
}

func (m methoder) m() string {
	return ""
}

func (a *Slice) Contains(el T) bool { return a.IndexOf(el) >= 0 }

func (a Slice) AppendUniq(el T) Slice {
	if a.IndexOf(el) < 0 {
		return append(a, el)
	}
	return a
}

func Empty() Slice {
	return nil
}

func (a *Slice) m() []methoder {
	return nil
}

func Empty1() *Slice {
	return nil
}

func Empty2() []Slice {
	return nil
}

func Empty3() [][]Slice {
	return nil
}
