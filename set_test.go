package ds

import "testing"

type Num int

func (n Num) Equals(v interface{}) bool {
	k, ok := v.(Num)

	if ok {
		return k == n
	}

	j, ok := v.(int)

	if ok {
		return Num(j) == n
	}

	return false
}

func massAdd(fx func(...Equalers), v ...int) {
	for _, i := range v {
		fx(Num(i))
	}
}

func TestStringSet(t *testing.T) {
	vs := NewStringSet()

	if vs == nil {
		t.Fatal("Unable to create string set")
	}

	vs.Add("alex")

	_, got := vs.Get("alex")

	if !got {
		t.Fatal("failed to get alex from set")
	}
}

func TestSet(t *testing.T) {
	bs := UnSafeSet()

	if bs == nil {
		t.Fatal("Unable to create base set")
	}

	massAdd(bs.Push, 1, 3, 1, 67, 5)

	if len(bs) < 5 {
		t.Fatalf("Length of set is unequal, expecting 5 got %d", len(bs))
	}

	bs.Add(Num(20), 2)

	if len(bs) < 5 {
		t.Fatalf("Length of set is unequal, expecting 6 got %d", len(bs))
	}

}

func TestDupSet(t *testing.T) {
	bs := UnSafeSet()

	if bs == nil {
		t.Fatal("Unable to create base set")
	}

	massAdd(bs.Push, 1, 1, 1, 5, 1, 20)

	if len(bs) < 5 {
		t.Fatalf("Length of set is unequal, expecting > 5 %d", len(bs))
	}

	bs.Sanitize()
	bs.Remove(5)

	if len(bs) > 3 {
		t.Fatalf("Length of set is unequal, expecting > 3 %d", len(bs))
	}
}

func TestBaseSet(t *testing.T) {
	bs := SafeSet()

	if bs == nil {
		t.Fatal("Unable to create base set")
	}

	massAdd(bs.Push, 1, 3, 4, 67, 5)

	if bs.Length() < 5 {
		t.Fatalf("Length of set is unequal, expecting 5 got %d", bs.Length())
	}

	if !bs.Contains(3) {
		t.Fatal("3 is not in set")
	}

	if bs.Contains(6) {
		t.Fatal("5 is in set")
	}

	_, state := bs.Find(30)

	if state {
		t.Fatal("Unexpected value 30 in set")
	}

}
