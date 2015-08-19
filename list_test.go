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

	if itr.Next() != nil {
		t.Fatal("List iterator can not get next even though just created")
	}

	itr.Reset()

	for itr.Next() == nil {
		t.Logf("Forward-Value: %+v:", itr.Value())
	}

	itr.Reset()

	ptr, ok := itr.(MovableDeferIterator)
	if ok {
		for ptr.Previous() == nil {
			t.Logf("Backward-Value: %+v", ptr.Value())
		}
	}

	t.Log("Finished Iterator List")

	pack.Clear()
}
