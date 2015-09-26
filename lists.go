package ds

import (
	"fmt"
	"sync/atomic"

	"github.com/influx6/sequence"
)

//NodePush defines a function for setting node links
type NodePush func(*DeferNode) *DeferNode

// DeferNode represents a standard node meeting DeferNode requirements
type DeferNode struct {
	data interface{}
	list *DeferList
	next NodePush
	prev NodePush
}

// DeferList represents a sets of linkedlist meeting DeferList requirements
type DeferList struct {
	tail NodePush
	root NodePush
	size int64
}

//Reset provides a convenient method to nill the next,prev link and the list it belongs to
func (d *DeferNode) Reset() {
	d.Detach()
	d.Disown()
	d.data = nil
}

//ChangeValue changes the value of the node
func (d *DeferNode) ChangeValue(v interface{}) {
	d.data = v
}

//ResetNext provides a convenient method to nill the next link
func (d *DeferNode) ResetNext() {
	d.next = nil
}

//String returns the string representation of the data
func (d *DeferNode) String() string {
	return fmt.Sprintf("%+v", d.data)
}

//ResetPrevious provides a convenient method to nill the previous link
func (d *DeferNode) ResetPrevious() {
	d.prev = nil
}

//Equals define the equality of two nodes
func (d *DeferNode) Equals(n interface{}) bool {
	dx, ok := n.(*DeferNode)

	if !ok {
		// return d.Value() == n
		return false
	}

	return dx.Value() == d.Value()
}

//Value returns the internal value of this node
func (d *DeferNode) Value() interface{} {
	return d.data
}

//Close sends a cascading signal to all nodes to kill themselves
func (d *DeferNode) Close() {
	if d.prev == nil && d.next == nil {
		return
	}

	prev := d.Previous()
	nxt := d.Next()

	d.ResetNext()
	d.ResetPrevious()
	d.data = nil
	d.list = nil

	if prev != nil {
		prev.Close()
	}

	if nxt != nil {
		nxt.Close()
	}

}

//Detach removes disconnect this node from its next and previous and reconnects those
func (d *DeferNode) Detach() {
	prev := d.Previous()
	nxt := d.Next()

	if prev != nil {
		if nxt != nil {
			nxt.UsePrevious(prev)
			prev.UseNext(nxt)
		} else {
			prev.ResetNext()
		}
	}

	d.ResetNext()
	d.ResetPrevious()
	d.list.decrement()
}

//Disown removes this node from its list without breaking the chains
func (d *DeferNode) Disown() {
	d.list = nil
}

//ChangeList resets the lists of the node
func (d *DeferNode) ChangeList(l *DeferList) {
	if l == nil {
		return
	}

	d.Detach()
	d.Disown()

	d.list = l
}

//List returns the list this node is associted with if any
func (d *DeferNode) List() *DeferList {
	return d.list
}

//Next returns the next node linked to this if existing
func (d *DeferNode) Next() *DeferNode {
	if d.next == nil {
		return nil
	}
	return d.next(d)
}

//Previous returns the previous node linked to this if existing
func (d *DeferNode) Previous() *DeferNode {
	if d.prev == nil {
		return nil
	}
	return d.prev(d)
}

//UsePrevious provides a convenient function for setting prev
func (d *DeferNode) UsePrevious(nx *DeferNode) {
	if nx == nil {
		return
	}

	if d == nx {
		return
	}

	d.ChangePrevious(func(_ *DeferNode) *DeferNode {
		return nx
	})
}

//UseNext provides a convenient function for setting next
func (d *DeferNode) UseNext(nx *DeferNode) {
	if nx == nil {
		return
	}

	if d == nx {
		return
	}

	d.ChangeNext(func(_ *DeferNode) *DeferNode {
		return nx
	})
}

//ChangeNext allows the changing/setting of the next node
func (d *DeferNode) ChangeNext(nx NodePush) {
	if nx == nil {
		return
	}
	d.next = nx
}

//ChangePrevious allows the changing/setting of the prev node
func (d *DeferNode) ChangePrevious(nx NodePush) {
	if nx == nil {
		return
	}
	d.prev = nx
}

//NewDeferNode returns a new deffered node
func NewDeferNode(data interface{}, l *DeferList) *DeferNode {
	return &DeferNode{data: data, list: l}
}

//List returns a new *DeferList instance
func List(data ...interface{}) (l *DeferList) {
	l = &DeferList{}

	for _, v := range data {
		if n, ok := v.(*DeferNode); ok {
			l.Add(n)
		} else {
			l.AppendElement(v)
		}
	}

	return
}

//AppendElement adds up a new data to the list
func (d *DeferList) AppendElement(data interface{}) *DeferNode {
	nc := NewDeferNode(data, d)
	d.Add(nc)
	return nc
}

//PopTail removes the last element and resets the tail to the previous element or returns a error
func (d *DeferList) PopTail() *DeferNode {
	if d.Tail() == nil {
		return nil
	}

	nx := d.Tail()
	prnx := nx.Previous()

	if prnx == nil {
		d.shiftTail(nil)
	}

	if nx == d.Root() {
		d.shiftRoot(nil)
	} else {
		d.shiftTail(prnx)
	}

	nx.Detach()

	return nx
}

//PopRoot removes the first element and resets the tail to the previous element or returns a error
func (d *DeferList) PopRoot() *DeferNode {
	if d.Root() == nil {
		return nil
	}

	nx := d.Root()
	rx := d.Root().Next()

	if rx == nil {
		d.shiftRoot(nil)
	}

	if nx == d.Tail() {
		d.shiftTail(nil)
	} else {
		d.shiftRoot(rx)
	}

	nx.Detach()

	return nx
}

//PrependElement adds up a new data to the list
func (d *DeferList) PrependElement(data interface{}) *DeferNode {
	nc := d.AddBefore(d.Root(), data)
	d.shiftRoot(nc)
	return nc
}

func (d *DeferList) increment() {
	atomic.AddInt64(&d.size, 1)
}

func (d *DeferList) decrement() {
	if d.size > 0 {
		atomic.AddInt64(&d.size, -1)
	}
}

//Add adds up a new node to the list
func (d *DeferList) Add(r *DeferNode) {
	if r == nil {
		return
	}

	d.increment()

	if d.root == nil && d.tail == nil {
		r.ChangeList(d)
		d.shiftRoot(r)
		d.shiftTail(r)
		return
	}

	r.ChangeList(d)
	dt := d.Tail()

	dt.UseNext(r)
	r.UsePrevious(dt)

	d.shiftTail(r)

	return
}

//shiftRoot provides a convenient setter
func (d *DeferList) shiftRoot(t *DeferNode) {
	d.root = func(_ *DeferNode) *DeferNode {
		return t
	}
}

//shiftTail provides a convenient setter
func (d *DeferList) shiftTail(r *DeferNode) {
	d.tail = func(_ *DeferNode) *DeferNode {
		return r
	}
}

//Root returns the root of the node
func (d *DeferList) Root() *DeferNode {
	if d.root == nil {
		return nil
	}
	return d.root(nil)
}

//Tail returns the tail of the node
func (d *DeferList) Tail() *DeferNode {
	if d.tail == nil {
		return nil
	}
	return d.tail(nil)
}

//IsEmpty returns a false/true to indicate emptiness
func (d *DeferList) IsEmpty() bool {
	return d.root == nil && d.tail == nil
}

//Length returns the size of the list
func (d *DeferList) Length() int {
	return int(atomic.LoadInt64(&d.size))
}

//Iterator returns the iterator capable of iterating to this list
func (d *DeferList) Iterator() sequence.Iterable {
	return NewListIterator(d)
}

//Parent returns the root sequence
func (d *DeferList) Parent() sequence.Sequencable {
	return nil
}

//AddNodeBefore adds up a new node before a supplied nodeto the list
func (d *DeferList) AddNodeBefore(f, n *DeferNode) {
	if !d.Has(f) {
		return
	}

	if !d.Has(n) {
		return
	}

	// nx := f.Next()

	f.UsePrevious(n)
	n.UseNext(f)

	prev := f.Previous()
	if prev != nil {
		n.UsePrevious(prev)
		prev.UseNext(n)
	}

	if d.Root() == f {
		d.shiftRoot(n)
	}
}

//AddBefore adds up a new node before a supplied nodeto the list
func (d *DeferList) AddBefore(f *DeferNode, r interface{}) *DeferNode {
	if !d.Has(f) {
		return nil
	}

	defer d.increment()
	n := NewDeferNode(r, d)

	d.AddNodeBefore(f, n)
	return n
}

//AddNodeAfter adds up a new node after a supplied nodeto the list
func (d *DeferList) AddNodeAfter(f, n *DeferNode) {
	if !d.Has(f) {
		return
	}

	if !d.Has(n) {
		return
	}

	f.UseNext(n)
	n.UsePrevious(f)

	nxt := f.Next()
	if nxt != nil {
		n.UseNext(nxt)
		nxt.UsePrevious(n)
	}

	if d.Tail() == f {
		d.shiftTail(n)
	}
}

//Release empties the list and returns the root and tail
func (d *DeferList) Release() (*DeferNode, *DeferNode) {
	r, t := d.Root(), d.Tail()
	if r != nil {
		r.Disown()
	}
	if r != nil {
		t.Disown()
	}
	return r, t
}

//AddAfter adds up a new node after a supplied nodeto the list
func (d *DeferList) AddAfter(f *DeferNode, r interface{}) *DeferNode {
	if !d.Has(f) {
		return nil
	}

	defer d.increment()
	n := NewDeferNode(r, d)

	d.AddNodeAfter(f, n)
	return n
}

//PushList pushes a copy of all nodes in the supplied list to the tail of this list
func (d *DeferList) PushList(r *DeferList) {
	nx := r.Iterator()

	for nx.Next() == nil {
		nr, ok := nx.Value().(*DeferNode)

		if ok {
			d.AppendElement(nr.Value())
		}
	}
}

//PushBackList pushes a copy of all nodes in the supplied list to the root of this list
func (d *DeferList) PushBackList(r *DeferList) {
	nx := r.Iterator()

	for nx.Next() == nil {
		nr, ok := nx.Value().(*DeferNode)

		if ok {
			d.PrependElement(nr.Value())
		}
	}
}

//Has returns true/false if a particular node exist in list
func (d *DeferList) Has(f *DeferNode) bool {
	if f == nil {
		return false
	}

	return f.List() == d
}

//Delete removes this node if it exists in the list
func (d *DeferList) Delete(f *DeferNode) interface{} {
	if !d.Has(f) {
		return nil
	}

	defer f.Reset()

	return f.Value()
}

//Detach removes disconnect this node from its next and previous and reconnects those
func (d *DeferList) Detach(f *DeferNode) {
	if !d.Has(f) {
		return
	}

	f.Detach()
}

//MoveToRoot resets the internal iterator to the root
func (d *DeferList) MoveToRoot(n *DeferNode) {
	d.AddNodeAfter(d.Root(), n)
	d.shiftRoot(n)
}

//MoveToTail resets the internal iterator to the tail
func (d *DeferList) MoveToTail(n *DeferNode) {
	d.Add(n)
}

//Clear empties this list and clears the internal nodes and their connection
func (d *DeferList) Clear() {
	r, t := d.Root(), d.Tail()

	if r != nil {
		r.Close()
	}
	if t != nil {
		t.Close()
	}

	d.tail = nil
	d.root = nil
}

//DeferIterator provides and defines methods for defer iteratore
type DeferIterator interface {
	sequence.Iterable
	Reset2Tail()
	Reset2Root()
}

//MovableDeferIterator defines iterators that can move both forward and backward
type MovableDeferIterator interface {
	DeferIterator
	Previous() error
}

//NewListIterator returns a iterator for the *DeferList
func NewListIterator(l *DeferList) *DeferListIterator {
	return &DeferListIterator{list: l}
}

//NewListIteratorAt returns a iterator for the *DeferList
func NewListIteratorAt(l *DeferList, d *DeferNode) (*DeferListIterator, error) {
	if !l.Has(d) {
		return nil, ErrBadNode
	}

	return &DeferListIterator{list: l, current: d}, nil
}

// DeferListIterator provides an iterator for DeferredList
type DeferListIterator struct {
	list    *DeferList
	current *DeferNode
	state   int64
}

//Length returns the length of the list
func (lx *DeferListIterator) Length() int {
	return lx.list.Length()
}

//Clone the iterator for a new one
func (lx *DeferListIterator) Clone() sequence.Iterable {
	return NewListIterator(lx.list)
}

//Reset defines the means to reset the iterator
func (lx *DeferListIterator) Reset() {
	atomic.StoreInt64(&lx.state, 0)
	lx.current = nil
}

//Reset2Root defines the means to reset the iterator to the root node
func (lx *DeferListIterator) Reset2Root() {
	lx.Reset()
	lx.current = lx.list.Root()
}

//Reset2Tail defines the means to reset the iterator to the tail node
func (lx *DeferListIterator) Reset2Tail() {
	lx.Reset()
	lx.current = lx.list.Tail()
}

//Value returns the data of current node as key
func (lx *DeferListIterator) Value() interface{} {
	v := lx.current

	if v == nil {
		return nil
	}

	return v.Value()
}

//Key returns the node as key
func (lx *DeferListIterator) Key() interface{} {
	return lx.current
}

//Previous defines the decrement for the iterator
func (lx *DeferListIterator) Previous() error {
	state := atomic.LoadInt64(&lx.state)

	if lx.current == nil && state > 0 {
		return sequence.ErrBADINDEX
	}

	if lx.current == nil && state <= 0 {
		if lx.list.Tail() == nil {
			return sequence.ErrBADINDEX
		}
		lx.current = lx.list.Tail()
		atomic.StoreInt64(&lx.state, 1)
		return nil
	}

	nx := lx.current.Previous()

	lx.current = nx

	if nx == nil {
		return sequence.ErrENDINDEX
	}

	return nil
}

//Next defines the incrementer for the iterator
func (lx *DeferListIterator) Next() error {
	state := atomic.LoadInt64(&lx.state)

	if lx.current == nil && state > 0 {
		return sequence.ErrBADINDEX
	}

	if lx.current == nil && state <= 0 {
		if lx.list.Root() == nil {
			return sequence.ErrBADINDEX
		}
		lx.current = lx.list.Root()
		atomic.StoreInt64(&lx.state, 1)
		return nil
	}

	nx := lx.current.Next()

	lx.current = nx
	if nx == nil {
		return sequence.ErrENDINDEX
	}

	return nil
}
