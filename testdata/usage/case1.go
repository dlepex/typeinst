// +build !release

package usage

import (
	"github.com/dlepex/typeinst/testdata/g/maps"
	"github.com/dlepex/typeinst/testdata/g/slices/filter"
	"github.com/dlepex/typeinst/testdata/g/slices/indexof"
)

//go:generate typeinst
type _typeinst struct {
	Dict    func(K string, V [][][]struct{}) maps.Map
	BigTree func(K int64, V interface{}) maps.TreeMap
	Ints    func(T int) indexof.Slice
	Floats  func(T float64) (indexof.Slice, filter.Slice)
}
