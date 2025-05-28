package ds

type Trie struct {
	root *trieNode
}

type trieNode struct {
	value *string
	childs []*trieNode
}

func newTrieNode(value *string) *trieNode {
	node := trieNode{}
	node.value = value
	node.childs = make([]*trieNode, 128)

	return &node
}

func newTrie() Trie {
	data := Trie{}
	data.root = newTrieNode(nil)

	return data
}

func (t *Trie) Add(data string) *string {
	if data == "" {
		return t.root.value
	}

	var current *trieNode = t.root
	var next *trieNode

	for _, char := range data {
		next = current.childs[char]
		if next == nil {
			next = newTrieNode(nil)
		}

		current = next
	}

	next.value = &data

	return next.value
}

func (t *Trie) Lookup(data string) bool {
	if data == "" {
		return true
	}

	var current *trieNode = t.root
	var next *trieNode

	for _, char := range data {
		next = current.childs[char]
		if next == nil {
			return false
		}

		current = next
	}

	return true
}