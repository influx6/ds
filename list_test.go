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
		if itr.Value() == nil {
			t.Fatal("Forward iterator produced a nil")
		}
	}

	itr.Reset()

	ptr, ok := itr.(MovableDeferIterator)
	if ok {
		for ptr.Previous() == nil {
			if ptr.Value() == nil {
				t.Fatal("Reverse iterator produced a nil")
			}
		}
	}

	pack.Clear()
}
