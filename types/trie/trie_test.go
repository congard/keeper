package trie

import (
	"testing"
)

func TestInsertAndSubtrie(t *testing.T) {
	trie := New[string, int]()

	// Insert a path and verify subtrie exists
	subtrie, isNew := trie.Insert("a", "b", "c")
	if !isNew {
		t.Fatal("expected new subtrie")
	}
	if subtrie == nil {
		t.Fatal("expected non-nil subtrie")
	}

	// Subtrie should retrieve the same node
	retrieved := trie.Subtrie("a", "b", "c")
	if retrieved != subtrie {
		t.Fatal("Subtrie should return inserted node")
	}

	// Intermediate nodes should exist
	if trie.Subtrie("a") == nil {
		t.Fatal("intermediate node 'a' should exist")
	}
	if trie.Subtrie("a", "b") == nil {
		t.Fatal("intermediate node 'b' should exist")
	}
}

func TestInsertExistingPathReturnsSameNode(t *testing.T) {
	trie := New[string, int]()

	subtrie1, isNew1 := trie.Insert("x", "y")
	if !isNew1 {
		t.Fatal("first insert should be new")
	}

	subtrie2, isNew2 := trie.Insert("x", "y")
	if isNew2 {
		t.Fatal("second insert should not be new")
	}
	if subtrie1 != subtrie2 {
		t.Fatal("re-inserting same path should return same node")
	}
}

func TestSubtrieNonExistentReturnsNil(t *testing.T) {
	trie := New[string, int]()
	trie.Insert("a", "b")

	if trie.Subtrie("a", "c") != nil {
		t.Fatal("non-existent path should return nil")
	}
	if trie.Subtrie("x") != nil {
		t.Fatal("non-existent root key should return nil")
	}
}

func TestSetValueAndHasValue(t *testing.T) {
	trie := New[string, int]()

	// Set value at path
	trie.SetValue(42, "a", "b", "c")

	// HasValue should return true for exact path
	if !trie.HasValue("a", "b", "c") {
		t.Fatal("HasValue should be true for set path")
	}

	// HasValue should be false for prefix paths
	if trie.HasValue("a") {
		t.Fatal("HasValue should be false for prefix without value")
	}
	if trie.HasValue("a", "b") {
		t.Fatal("HasValue should be false for intermediate node without value")
	}

	// HasValue should be false for non-existent paths
	if trie.HasValue("a", "b", "d") {
		t.Fatal("HasValue should be false for non-existent path")
	}
}

func TestValueRetrieval(t *testing.T) {
	trie := New[string, int]()
	trie.SetValue(100, "key")

	val, ok := trie.Value("key")
	if !ok {
		t.Fatal("Value should return ok=true for existing key")
	}
	if val != 100 {
		t.Fatalf("Value = %d, want 100", val)
	}

	// Non-existent key
	val, ok = trie.Value("missing")
	if ok {
		t.Fatal("Value should return ok=false for missing key")
	}
	if val != 0 {
		t.Fatalf("Value = %d, want 0 for missing key", val)
	}
}

func TestValueAtEmptyPath(t *testing.T) {
	trie := New[string, int]()
	trie.SetValue(999)

	val, ok := trie.Value()
	if !ok {
		t.Fatal("Value at root should work")
	}
	if val != 999 {
		t.Fatalf("Value = %d, want 999", val)
	}
}

func TestOverwriteValue(t *testing.T) {
	trie := New[string, int]()
	trie.SetValue(1, "key")
	trie.SetValue(2, "key")

	val, ok := trie.Value("key")
	if !ok || val != 2 {
		t.Fatalf("Value = %d, ok=%v, want 2, true", val, ok)
	}
}

func TestCutRemovesSubtrie(t *testing.T) {
	trie := New[string, int]()
	trie.Insert("a", "b", "c")
	trie.SetValue(10, "a", "b")
	trie.SetValue(20, "a", "b", "c")

	// Cut at "a", "b" removes the "b" subtrie from "a"
	cut := trie.Cut("a", "b")
	if cut == nil {
		t.Fatal("Cut should return removed subtrie")
	}

	// Cut subtrie retains its value and children
	if !cut.HasValue() {
		t.Fatal("cut subtrie should retain its own value")
	}
	val, _ := cut.Value()
	if val != 10 {
		t.Fatalf("cut subtrie value = %d, want 10", val)
	}
	if !cut.HasValue("c") {
		t.Fatal("cut subtrie should retain its children")
	}

	// Original trie no longer has the cut path
	if trie.Subtrie("a", "b") != nil {
		t.Fatal("original trie should not have cut path")
	}
	if trie.HasValue("a", "b") {
		t.Fatal("original trie should not have cut path value")
	}
	if trie.HasValue("a", "b", "c") {
		t.Fatal("original trie should not have cut path children")
	}

	// But "a" still exists
	if trie.Subtrie("a") == nil {
		t.Fatal("original trie should retain parent node")
	}
}

func TestCutNonExistentReturnsNil(t *testing.T) {
	trie := New[string, int]()
	trie.Insert("a", "b")

	if trie.Cut("a", "c") != nil {
		t.Fatal("Cut non-existent should return nil")
	}
	if trie.Cut("x") != nil {
		t.Fatal("Cut non-existent root should return nil")
	}
}

func TestInsertWithIntKeys(t *testing.T) {
	trie := New[int, string]()
	trie.SetValue("hello", 1, 2, 3)

	val, ok := trie.Value(1, 2, 3)
	if !ok || val != "hello" {
		t.Fatalf("Value = %q, ok=%v, want \"hello\", true", val, ok)
	}
}

func TestEmptyPathOperations(t *testing.T) {
	trie := New[string, int]()

	// Insert with empty path returns root
	subtrie, isNew := trie.Insert()
	if isNew {
		t.Fatal("Insert with empty path should not be new")
	}
	if subtrie != trie {
		t.Fatal("Insert with empty path should return root")
	}

	// Subtrie with empty path returns root
	if trie.Subtrie() != trie {
		t.Fatal("Subtrie with empty path should return root")
	}

	// Cut with empty path returns nil
	if trie.Cut() != nil {
		t.Fatal("Cut with empty path should return nil")
	}
}
