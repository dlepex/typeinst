package indexof

type T = interface{} //typeinst: typevar

type Slice []T

func NewSlice() Slice {
	return nil
}

func WithCap(a int) Slice {
	f := NewSlice
	return f()
}

func (a Slice) IndexOf(el T) int {
	for i, v := range a {
		if v == el {
			return i
		}
	}
	return -1
}

func (a Slice) WithCap() {

}

func (a *Slice) Contains(el T) bool { return a.IndexOf(el) >= 0 }

func (a Slice) AppendUniq(el T) Slice {
	if a.IndexOf(el) < 0 {
		return append(a, el)
	}
	return a
}

func Empty() {

}
