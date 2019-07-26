// nolint
package set

type E = interface{}

type Set map[E]struct{}

func New() Set {
	return make(map[E]struct{})
}

func (set Set) Add(elem E) {
	set[elem] = struct{}{}
}

func (set Set) Has(elem E) bool {
	_, has := set[elem]
	return has
}

type SetToSlice struct{}

func (_ SetToSlice) F(set Set, slice []E) []E {
	for e, _ := range set {
		slice = append(slice, e)
	}
	return slice
}
