package maps

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
