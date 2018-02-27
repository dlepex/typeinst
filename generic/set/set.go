package set

type T = interface{}

type Set map[T]struct{}

func NewSet() Set {
	return make(map[T]struct{})
}

func (set Set) Add(v T) {
	set[v] = struct{}{}
}

func (set Set) Has(v T) bool {
	_, has := set[v]
	return has
}
