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
