package optimizer

import (
	"math/rand"
	"testing"
)

func TestSuggestQuotasDeterministic(t *testing.T) {
	b := NewBandit()
	for i := 0; i < 10; i++ {
		b.Update("A", 1)
	}
	for i := 0; i < 10; i++ {
		b.Update("B", 0)
	}
	for i := 0; i < 3; i++ {
		b.Ensure("C")
	}

	q1 := SuggestQuotas(rand.New(rand.NewSource(7)), b, 30)
	q2 := SuggestQuotas(rand.New(rand.NewSource(7)), b, 30)

	if q1["A"] != q2["A"] || q1["B"] != q2["B"] || q1["C"] != q2["C"] {
		t.Fatalf("not deterministic q1=%v q2=%v", q1, q2)
	}
	if q1["A"] <= q1["B"] {
		t.Fatalf("expected A to get >= B slots q=%v", q1)
	}
}
