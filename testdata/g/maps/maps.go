package maps

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

type Node struct {
	key  K
	val  V
	h    int
	next **Node
}

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

}
