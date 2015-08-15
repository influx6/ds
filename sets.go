package ds

import (
	"sync"
	"sync/atomic"
)

//NewNodeSet returns the set for nodes
func NewNodeSet() *NodeSet {
	return &NodeSet{
		set: SafeSet(),
	}
}

//NewDeferNodeSet returns the set for nodes
func NewDeferNodeSet() *DeferNodeSet {
	return &DeferNodeSet{
		set: SafeSet(),
	}
}

//Length returns the size of the set
func (n *NodeSet) Length() int {
	return n.set.Length()
}

//GetIndex returns the node at this index
func (n *NodeSet) GetIndex(ind int) (Nodes, error) {
	if ind < 0 {
		ind = n.set.Length() - ind
	}

	if ind >= n.set.Length() {
		return nil, ErrBadIndex
	}

	return n.set.Get(ind).(Nodes), nil
}

//FirstNode returns the first node
func (n *NodeSet) FirstNode() (Nodes, error) {
	return n.GetIndex(0)
}

//AllNodes return the internal nodes
func (n *NodeSet) AllNodes() []Nodes {
	nodes := []Nodes{}

	n.set.Each(func(v Equalers, _ int, _ func()) {
		nodes = append(nodes, v.(Nodes))
	})

	return nodes
}

//LastNode returns the first node
func (n *NodeSet) LastNode() (Nodes, error) {
	return n.GetIndex(n.Length() - 1)
}

//RemoveNode adds a new node into the list
func (n *NodeSet) RemoveNode(data Nodes) {
	n.set.Remove(data)
}

//AddNode adds a new node into the list
func (n *NodeSet) AddNode(data Nodes) {
	atomic.StoreInt64(&n.dirty, 1)
	n.set.Push(data)
}

//EachNode calls the internal set Each method
func (n *NodeSet) EachNode(fx func(Nodes)) {
	if fx == nil {
		return
	}
	n.Each(func(nx Nodes, _ int, _ func()) {
		fx(nx)
	})
}

//Each calls the internal set Each method
func (n *NodeSet) Each(fx func(Nodes, int, func())) {
	if fx == nil {
		return
	}
	n.set.Each(func(v Equalers, k int, sm func()) {
		nx, ok := v.(Nodes)
		if ok {
			fx(nx, k, sm)
		}
	})
}

//GetNode return the node if found with the value
func (n *NodeSet) GetNode(data interface{}) (Nodes, bool) {

	if n.dirty > 0 {
		n.set.Sanitize()
		atomic.StoreInt64(&n.dirty, 0)
	}

	ind, ok := n.set.Find(data)

	if ok {
		pn, _ := n.set.Get(ind).(Nodes)
		return pn, ok
	}

	return nil, ok
}

//AllNodes return the internal nodes
func (n *DeferNodeSet) AllNodes() []*DeferNode {
	nodes := []*DeferNode{}

	n.set.Each(func(v Equalers, _ int, _ func()) {
		nodes = append(nodes, v.(*DeferNode))
	})

	return nodes
}

//Length returns the size of the set
func (n *DeferNodeSet) Length() int {
	return n.set.Length()
}

//RemoveNode adds a new node into the list
func (n *DeferNodeSet) RemoveNode(data *DeferNode) {
	n.set.Remove(data)
}

//GetIndex returns the node at this index
func (n *DeferNodeSet) GetIndex(ind int) (*DeferNode, error) {
	if ind <= 0 {
		ind = n.set.Length() - ind
	}
	if ind >= n.set.Length() {
		return nil, ErrBadIndex
	}
	return n.set.Get(ind).(*DeferNode), nil
}

//FirstNode returns the first node
func (n *DeferNodeSet) FirstNode() (*DeferNode, error) {
	return n.GetIndex(0)
}

//LastNode returns the first node
func (n *DeferNodeSet) LastNode() (*DeferNode, error) {
	return n.GetIndex(n.Length() - 1)
}

//AddNode adds a new node into the list
func (n *DeferNodeSet) AddNode(data *DeferNode) {
	defer atomic.StoreInt64(&n.dirty, 1)
	n.set.Push(data)
}

//EachNode calls the internal set Each method
func (n *DeferNodeSet) EachNode(fx func(*DeferNode)) {
	if fx == nil {
		return
	}
	n.Each(func(nx *DeferNode, _ int, _ func()) {
		fx(nx)
	})
}

//Each calls the internal set Each method
func (n *DeferNodeSet) Each(fx func(*DeferNode, int, func())) {
	n.set.Each(func(v Equalers, k int, sm func()) {
		nx, ok := v.(*DeferNode)
		if ok {
			fx(nx, k, sm)
		}
	})
}

//GetNode return the node if found with the value
func (n *DeferNodeSet) GetNode(data interface{}) (*DeferNode, bool) {

	if n.dirty > 0 {
		n.set.Sanitize()
		atomic.StoreInt64(&n.dirty, 0)
	}

	ind, ok := n.set.Find(data)

	if ok {
		pn, _ := n.set.Get(ind).(*DeferNode)
		return pn, ok
	}

	return nil, ok
}

//SafeSet returns a new BaseSet
func SafeSet() *baseset {
	return &baseset{UnSafeSet(), new(sync.RWMutex)}
}

//UnSafeSet returns a new BaseSet
func UnSafeSet() set {
	return make(set, 0)
}

//Length returns the length of the set
func (b *baseset) Length() int {
	b.rw.RLock()
	sz := len(b.set)
	b.rw.RUnlock()
	return sz
}

//Add adds an item into the set return a bool wether succesful or not
func (b *baseset) Add(e Equalers, pos int) (int, bool) {
	b.rw.Lock()
	ind, state := b.set.Add(e, pos)
	b.rw.Unlock()
	return ind, state
}

//Add adds an item into the set return a bool wether succesful or not
func (b *set) Add(e Equalers, pos int) (int, bool) {
	if pos <= -1 || len(*b) <= pos {
		*b = append(*b, e)
		return len(*b), false
	}

	p1 := (*b)[:pos+1]

	tmp := make(set, 0)
	tmp = append(tmp, (*b)[pos:]...)

	(*b)[pos] = e

	*b = append(p1, tmp...)

	tmp = nil

	return pos, true
}

//Push adds items into the list
func (b *set) Push(e ...Equalers) {
	for _, v := range e {
		_, _ = b.Add(v, -1)
	}
}

//Push adds items into the list
func (b *baseset) Push(e ...Equalers) {
	b.rw.Lock()
	b.set.Push(e...)
	b.rw.Unlock()
}

//Each iterates all set data using a callback
func (b *baseset) Each(fx func(Equalers, int, func())) {
	if fx == nil {
		return
	}

	kill := false
	for k, v := range b.set {
		if kill {
			break
		}
		fx(v, k, func() { kill = true })
	}
}

//Get gets the items at the index
func (b *baseset) Get(e int) Equalers {
	b.rw.RLock()
	defer b.rw.RUnlock()
	return b.set[e]
}

//Sanitize gets the items at the index
func (b *baseset) Sanitize() {
	b.rw.Lock()
	b.set.Sanitize()
	b.rw.Unlock()
}

//Find gets the items at the index
func (b *baseset) Find(e interface{}) (int, bool) {
	b.rw.RLock()
	ci, cs := b.set.Find(e)
	b.rw.RUnlock()
	return ci, cs
}

//Contains gets the items at the index
func (b *baseset) Contains(e interface{}) bool {
	b.rw.RLock()
	cs := b.set.Contains(e)
	b.rw.RUnlock()
	return cs
}

//Contains returns true if the value exists in set
func (b *set) Contains(g interface{}) bool {
	_, s := b.Find(g)
	return s
}

//Find returns (index,bool) to indicate the position and if indeed the value exists else returns the last index and a false value
func (b *set) Find(g interface{}) (int, bool) {
	for n, v := range *b {
		if !v.Equals(g) {
			continue
		}
		return n, true
	}

	return len(*b), false
}

//Remove deletes this value from the set
func (b *baseset) Remove(g interface{}) Equalers {
	b.rw.Lock()
	s := b.set.Remove(g)
	b.rw.Unlock()
	return s
}

//Remove deletes this value from the set
func (b *set) Remove(g interface{}) Equalers {
	nm, state := b.Find(g)

	if state {
		tmp := (*b)[nm]

		*b = append((*b)[0:nm], (*b)[nm+1:]...)
		return tmp
	}

	return nil
}

//Sanitize removes all duplicates
func (b *set) Sanitize() {
	sz := len(*b) - 1
	for i := 0; i < sz; i++ {
		for j := i + 1; j <= sz; j++ {
			if (*b)[i].Equals((*b)[j]) {
				(*b)[j] = (*b)[sz]
				(*b) = (*b)[0:sz]
				sz--
				j--
			}
		}
	}
}
