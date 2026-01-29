package optimizer

import (
	"math/rand"
	"testing"
)

func TestBanditUpdateAndSuggest(t *testing.T) {
	b := NewBandit()
	// Arm A wins, arm B loses.
	for i := 0; i < 10; i++ {
		b.Update("a", 1)
		b.Update("b", 0)
	}
	r := rand.New(rand.NewSource(7))
	alloc, err := b.SuggestAllocation(r, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(alloc) != 1 || alloc[0].Key != "a" {
		t.Fatalf("expected top arm 'a', got=%v", alloc)
	}
}

