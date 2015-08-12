package ds

import "testing"

func TestListIterator(t *testing.T) {
	pack := List(1, 2, 3)

	if pack == nil {
		t.Fatal("Error occured creating list")
	}

	itr := pack.Iterator()

	if itr == nil {
		t.Fatal("Error occured creating list iterator")
	}

	if !itr.HasNext() {
		t.Fatal("List iterator can not get next even though just created")
	}

	for itr.HasNext() {
		t.Logf("Forward-Value: %+v:", itr.Value())
		itr.Next()
	}

	itr.Reset()

	ptr, ok := itr.(MovableDeferIterator)
	if ok {
		for ptr.HasPrevious() {
			t.Logf("Backward-Value: %+v", ptr.Value())
			ptr.Previous()
		}
	}

	t.Log("Finished Iterator List")

	pack.Clear()

	t.Log("Log:", pack)
}
