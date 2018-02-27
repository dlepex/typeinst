package filter

import (
	fmt1 "fmt"
	"log"
)

type T = interface{} //typeinst: typevar

type Slice []T

func (a Slice) FilterInplace(f func(T) bool) Slice {
	var d []T = a[:0]
	for _, v := range a {
		if f(v) {
			d = append(d, v)
		}
	}
	if len(d) == 0 {
		return nil
	}
	return d
}

func (a Slice) FilterInplaceZ(f func(T) bool, zero T) Slice {
	log.Printf("Hello")
	fmt1.Printf("Hello")
	d := a.FilterInplace(f)
	k := len(d)
	if k != 0 && k != len(a) {
		tail := a[k:]
		for i, _ := range tail {
			tail[i] = zero
		}
	}
	return d
}
