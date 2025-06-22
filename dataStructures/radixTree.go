package datastructures

type radixTreeNode[T any] struct {
	data               *T
	value              string
	childatastructures []*radixTreeNode[T]
}

type RadixTree[T any] struct {
	head *radixTreeNode[T]
}

func (rt *RadixTree[T]) Add(value string, data T) {
	current := &rt.head

	for *current != nil {
		if len((*current).childatastructures) == 0 {
			(*current).childatastructures = append((*current).childatastructures, nil)
			current = &(*current).childatastructures[0]
		} else {
			currentChild := 0
			for currentChild < len((*current).childatastructures) {
				prefixMatch := comparePrefixWithValue((*current).childatastructures[currentChild].value, value)
				if prefixMatch == 0 {
					currentChild++
					continue
				}

				if len((*current).childatastructures[currentChild].value) < prefixMatch {
					// TODO: split prefix
					break
				} else {
					// TODO: replace value
					break
				}
			}
		}
	}

	*current = newRadixNode(value, data)

}

func newRadixNode[T any](value string, data T) *radixTreeNode[T] {
	node := radixTreeNode[T]{}
	node.data = &data
	node.value = value
	node.childatastructures = make([]*radixTreeNode[T], 0)

	return &node
}

func comparePrefixWithValue(value string, prefix string) int {
	index := int(0)
	upperLimit := int(min(len(value), len(prefix)))

	for value[index] == prefix[index] && index < upperLimit {
		index++
	}

	return index
}
