package ds

import (
	"fmt"
	"sync/atomic"

	"github.com/influx6/sequence"
)

var (
	defaultVisit = func(n Nodes, visited bool) bool {
		if visited {
			return false
		}
		return true
	}
)

//BFPreOrderDirective provides a copy of a breadth-first pre order rule
func BFPreOrderDirective(d int, v VisitCaller) *TransversalDirective {
	return MakeTransversalDirective(d, BFPreOrder, v)
}

//BFPostOrderDirective provides a copy of a breath-first pre order rule
func BFPostOrderDirective(d int, v VisitCaller) *TransversalDirective {
	return MakeTransversalDirective(d, BFPostOrder, v)
}

//DFPreOrderDirective provides a copy of a depth-first pre order rule
func DFPreOrderDirective(d int, v VisitCaller) *TransversalDirective {
	return MakeTransversalDirective(d, DFPreOrder, v)
}

//DFPostOrderDirective provides a copy of a depth-first pre order rule
func DFPostOrderDirective(d int, v VisitCaller) *TransversalDirective {
	return MakeTransversalDirective(d, DFPostOrder, v)
}

//Search returns a GraphProc for depth-first
func Search(fx GraphHandlers, dir *TransversalDirective) (*GraphProc, error) {

	var verso *Transversor

	switch dir.Order {
	case DFPostOrder:
		verso = DepthFirstPostOrder(dir)
	case DFPreOrder:
		verso = DepthFirstPreOrder(dir)
	case BFPostOrder:
		verso = BreadthFirstPostOrder(dir)
	case BFPreOrder:
		verso = BreadthFirstPreOrder(dir)
	default:
		return nil, fmt.Errorf("Unknown Transversal Order %s", dir.Order)
	}

	return &GraphProc{
		trans: verso,
		fx:    fx,
	}, nil
}

//Use sets the node for transversal
func (p *GraphProc) Use(n Nodes) {
	if n == nil {
		return
	}

	gr := n.Graph()
	if gr == nil {
		return
	}

	p.g = gr
	p.trans.Use(n)
}

//Unvisited calls the internal transversal unvisited caller
func (p *GraphProc) Unvisited() []Nodes {
	if p.g == nil {
		return nil
	}
	return p.trans.Unvisited(p.g)
}

//Reset calls the internal transversal reset caller
func (p *GraphProc) Reset() {
	p.g = nil
	p.trans.Reset()
}

//WalkDepth returns the current depth of Transversor
func (p *GraphProc) WalkDepth() int {
	return p.trans.WalkDepth()
}

//Next calls the internal transversal next caller
func (p *GraphProc) Next() error {
	if err := p.trans.Next(); err != nil {
		return err
	}
	return p.fx(p.trans.Node(), p.trans.Key())
}

//MakeTransversalDirective creates a transversal directive
func MakeTransversalDirective(depth int, order TransversalOrder, visit VisitCaller) *TransversalDirective {
	if visit == nil {
		visit = defaultVisit
	}
	return &TransversalDirective{
		Depth:    depth,
		Order:    order,
		Revisits: visit,
	}
}

//MakeTransversor returns a default Transversor
func MakeTransversor(dir *TransversalDirective) *Transversor {
	core := &Transversor{
		directive: dir,
		visited:   VisitMaps(),
		keys:      make(map[Nodes]*Socket),
	}

	return core
}

//Use sets the node to tranversal from
func (t *Transversor) Use(n Nodes) {
	t.from = n
}

//Unvisited returns the unvisited nodes
func (t *Transversor) Unvisited(graph Graphs) []Nodes {
	return UnvisitedUtil(graph, t.visited)
}

//WalkDepth returns the current depth of Transversor
func (t *Transversor) WalkDepth() int {
	return int(t.walkdepth)
}

//Key returns the Socket of the node if its not a root node
func (t *Transversor) Key() *Socket {
	sc, ok := t.keys[t.current]
	if !ok {
		return nil
	}
	return sc
}

//Next calls the internal next
func (t *Transversor) Next() error {
	if t.from == nil {
		return ErrBadNode
	}

	if t.next != nil {
		return t.next()
	}

	return ErrBadIterator
}

//Node returns the current  node
func (t *Transversor) Node() Nodes {
	return t.current
}

//Reset resets the tranveror
func (t *Transversor) Reset() {
	t.from = nil
	t.started = 0
	t.walkdepth = 0
	t.current = nil
	t.keys = make(map[Nodes]*Socket)
	t.visited.Reset()
	if t.reset != nil {
		t.reset()
	}
}

//DepthFirstPreOrder returns a depth first search provider
func DepthFirstPreOrder(directive *TransversalDirective) (t *Transversor) {
	t = MakeTransversor(directive)

	var cache = NewCache()
	unlocked := true

	t.reset = func() {
		cache.Reset()
	}

	t.next = func() error {
		if cache.Length() <= 0 && t.started > 0 {
			t.current = nil
			return sequence.ErrBADINDEX
		}

		if t.started <= 0 {
			atomic.StoreInt64(&t.started, 1)
			unlocked = false
			cache.AddCache(t.from)
		}

		cur, err := cache.LastCache()

		if err != nil {
			t.current = nil
			return err
		}

		node, itr := cur.Node, cur.Itr

		if t.directive.Revisits(node, t.visited.Valid(node)) {
			t.visited.Add(node)
			t.current = node
			return nil
		}

		err = itr.Next()

		if err != nil {
			atomic.AddInt64(&t.walkdepth, -1)
			cache.Uncache()
			return t.Next()
		}

		cursoc, ok := itr.Value().(*Socket)

		if !ok {
			t.current = nil
			return ErrBadEdgeType
		}

		node = cursoc.To

		if !t.directive.Revisits(node, t.visited.Valid(node)) {
			return t.Next()
		}

		t.keys[node] = cursoc
		cache.AddCache(node)

		atomic.AddInt64(&t.walkdepth, 1)

		return t.Next()
	}

	return
}

//DepthFirstPostOrder returns a depth first search provider
func DepthFirstPostOrder(directive *TransversalDirective) (t *Transversor) {
	t = MakeTransversor(directive)

	var cache = NewCache()
	var curnode *NodeCache
	var err error

	t.reset = func() {
		cache.Reset()
	}

	t.next = func() error {
		if cache.Length() <= 0 && t.started > 0 {
			return ErrBadIndex
		}

		if t.started <= 0 {
			atomic.StoreInt64(&t.started, 1)
			cache.AddCache(t.from)
			t.visited.Add(t.from)
		}

		// t.key = nil
		curnode, err = cache.LastCache()

		if err != nil {
			t.current = nil
			return ErrBadIndex
		}

		cur := curnode.Itr
		node := curnode.Node

		if err := cur.Next(); err != nil {
			// cache.UncacheRight()
			if t.directive.Revisits(node, t.visited.Valid(node)) {
				cache.AddCache(node)
				if t.walkdepth > 0 {
					atomic.AddInt64(&t.walkdepth, -1)
				}
				t.visited.Add(node)
				return t.Next()
			}

			if t.walkdepth > 0 {
				atomic.AddInt64(&t.walkdepth, -1)
			}

			cache.Uncache()
			// t.visited.Add(node)
			t.current = node
			return nil
		}

		curnode, ok := cur.Value().(*Socket)

		if !ok {
			t.current = nil
			return ErrBadEdgeType
		}

		// if _, ok := socks[curnode.To]; !ok {
		// 	socks[curnode.To] = curnode
		// 	log.Printf("Sockets: %+s", socks)
		// }

		co := curnode.To

		atomic.AddInt64(&t.walkdepth, 1)
		if !t.directive.Revisits(co, t.visited.Valid(co)) {
			return t.Next()
		}

		t.keys[co] = curnode

		cache.AddCache(co)
		t.visited.Add(co)
		// t.current = co
		return t.Next()
	}

	return
}

//BreadthFirstPreOrder returns a depth first search provider
func BreadthFirstPreOrder(directive *TransversalDirective) (t *Transversor) {
	t = MakeTransversor(directive)

	unlocked := true
	var cache = NewCache()

	t.reset = func() {
		cache.Reset()
	}

	t.next = func() error {
		if cache.Length() <= 0 && t.started > 0 {
			return sequence.ErrBADINDEX
		}

		if cache.Length() <= 0 {
			atomic.StoreInt64(&t.started, 1)
			cache.AddCache(t.from)
			t.visited.Add(t.from)
			t.current = t.from
			return nil
		}

		cur, err := cache.FirstCache()

		if err != nil {
			t.current = nil
			return err
		}

		node, itr := cur.Node, cur.Itr

		if t.directive.Revisits(node, t.visited.Valid(node)) {
			cache.AddCache(node)
			return t.Next()
		}

		if err := itr.Next(); err != nil {
			unlocked = true
			cache.UncacheRight()
			return t.Next()
		}

		if unlocked {
			atomic.AddInt64(&t.walkdepth, 1)
			unlocked = false
		}

		soc, ok := itr.Value().(*Socket)

		if !ok {
			t.current = nil
			return ErrBadEdgeType
		}

		no := soc.To

		if !t.directive.Revisits(no, t.visited.Valid(no)) {
			return t.Next()
		}

		t.keys[no] = soc

		t.visited.Add(no)
		cache.AddCache(no)
		t.current = no
		return nil
	}

	return
}

//BreadthFirstPostOrder returns a depth first search provider
func BreadthFirstPostOrder(directive *TransversalDirective) (t *Transversor) {
	t = MakeTransversor(directive)

	var cache = NewCache()
	var curnode *NodeCache
	var err error
	depths := make(map[Nodes]int64)

	t.reset = func() {
		cache.Reset()
	}

	t.next = func() error {
		if cache.Length() <= 0 && t.started > 0 {
			depths = nil
			return ErrBadIndex
		}

		if t.started <= 0 {
			atomic.StoreInt64(&t.started, 1)
			depths[t.from] = int64(0)
			cache.AddCache(t.from)
		}

		curnode, err = cache.FirstCache()

		if err != nil {
			t.current = nil
			return ErrBadIndex
		}

		cur := curnode.Itr
		node := curnode.Node

		if err := cur.Next(); err != nil {
			cache.UncacheRight()

			if !t.directive.Revisits(node, t.visited.Valid(node)) {
				return t.Next()
			}

			atomic.StoreInt64(&t.walkdepth, depths[node])
			t.visited.Add(node)
			t.current = node
			return nil
		}

		curnode, ok := cur.Value().(*Socket)

		if !ok {
			t.current = nil
			return ErrBadEdgeType
		}

		co := curnode.To

		if !t.directive.Revisits(co, t.visited.Valid(co)) {
			return t.Next()
		}

		t.keys[co] = curnode

		if _, ok := depths[co]; !ok {
			dp := depths[node]
			depths[co] = dp + 1
			atomic.StoreInt64(&t.walkdepth, depths[co])
		}

		atomic.StoreInt64(&t.walkdepth, depths[co])
		cache.AddCache(co)
		t.visited.Add(co)
		t.current = co
		return nil
	}

	return
}
