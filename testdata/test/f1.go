package test

type Tux int

func (t Tux) Test1(sl Slice) Slice {
	var x Slice = sl
	return x
}
