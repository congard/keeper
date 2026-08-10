package trie

type Trie[K comparable, V any] struct {
	subtries map[K]*Trie[K, V]
	value    *V
}

func New[K comparable, V any]() *Trie[K, V] {
	return &Trie[K, V]{
		subtries: make(map[K]*Trie[K, V]),
		value:    nil,
	}
}

func (trie *Trie[K, V]) Subtrie(path ...K) *Trie[K, V] {
	if len(path) == 0 {
		return trie
	}

	if subtrie, ok := trie.subtries[path[0]]; ok {
		return subtrie.Subtrie(path[1:]...)
	}

	return nil
}

func (trie *Trie[K, V]) Insert(path ...K) (subtrie *Trie[K, V], isNew bool) {
	if len(path) == 0 {
		subtrie = trie
		isNew = false
		return
	}

	if subtrie, ok := trie.subtries[path[0]]; ok {
		return subtrie.Insert(path[1:]...)
	}

	subtrie = New[K, V]()
	trie.subtries[path[0]] = subtrie

	if len(path) == 1 {
		isNew = true
		return
	}

	return subtrie.Insert(path[1:]...)
}

func (trie *Trie[K, V]) Cut(path ...K) *Trie[K, V] {
	if len(path) == 0 {
		return nil
	}

	if subtrie, ok := trie.subtries[path[0]]; ok {
		if len(path) == 1 {
			delete(trie.subtries, path[0])
			return subtrie
		}
		return subtrie.Cut(path[1:]...)
	}

	return nil
}

func (trie *Trie[K, V]) SetValue(value V, path ...K) *Trie[K, V] {
	subtrie, _ := trie.Insert(path...)
	subtrie.value = &value
	return subtrie
}

func (trie *Trie[K, V]) HasValue(path ...K) bool {
	subtrie := trie.Subtrie(path...)
	return subtrie != nil && subtrie.value != nil
}

func (trie *Trie[K, V]) Value(path ...K) (value V, ok bool) {
	subtrie := trie.Subtrie(path...)
	if subtrie == nil || subtrie.value == nil {
		ok = false
		return
	}
	value = *subtrie.value
	ok = true
	return
}
