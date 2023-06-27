package synx

import "testing"

func TestMap(t *testing.T) {
	var m Map[int, string]

	m.Store(10, "apples")
	m.Store(20, "bananas")

	got, ok := m.Load(10)
	if !ok {
		t.Fatal()
	}
	if got != "apples" {
		t.Fatalf("have %s, want %s", got, "apples")
	}
}
