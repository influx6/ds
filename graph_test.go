package ds

import "testing"

func TestGraph(t *testing.T) {
	gs := NewGraph()

	if gs == nil {
		t.Fatal("Failed creating graph!")
	}

	gs.Add("alex", "john", "Block", "Date")

	if gs.Length() < 4 {
		t.Fatal("Graph does not contain appropriate elements!")
	}

	if !gs.Contains("alex") {
		t.Fatal("Graph does not contains element 'alex'")
	}

	_, state := gs.Bind("alex", "john", 20)

	if !state {
		t.Fatal("Unable to find 'alex' or 'john'")
	}

	if !gs.IsBound("alex", "john") {
		t.Fatal("'alex' is not bound to 'john'")
	}

	if gs.IsBound("john", "alex") {
		t.Fatal("'john' is not bound to 'alex'")
	}
}
