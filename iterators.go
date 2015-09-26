package ds

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/influx6/sequence"
)

var (
	// ErrNotFound is returned when no node is found matching criteria
	ErrNotFound = errors.New("Node not Found")

	defaultVisit = func(n Nodes, visited bool) bool {
		if visited {
			return false
		}
		return true
	}
	defaultHeuristic = func(n Nodes, soc *Socket) error {
		return nil
	}
	heumap = func(nx NodeOp) NodeOp {
		return func(n Nodes, soc *Socket) error {
			if soc == nil && n != nil {
				return nil
			}
			return nx(n, soc)
		}
	}
)

//TransversalOrder provides a order for TranversalDirective
type TransversalOrder string

const (
	//DFPreOrder represents depth first order of starting with the node then go to the child nodes
	DFPreOrder TransversalOrder = "depth-first-preorder"
	//DFPostOrder represents depth first order of starting with the node children then go to the node
	DFPostOrder TransversalOrder = "depth-first-postorder"
	//BFPreOrder represents breadth first order of starting with the node then go to the child nodes
	BFPreOrder TransversalOrder = "breadth-first-preorder"
	//BFPostOrder represents breadth first order of starting with the node children then go to the node
	BFPostOrder TransversalOrder = "breadth-first-postorder"
)

//VisitCaller provides a type for visit checks
type VisitCaller func(Nodes, bool) bool

//NodeOp provide a type for running on graph iterators
type NodeOp func(Nodes, *Socket) error

//NodeEval provides a evalutor format type
type NodeEval func(Nodes, *Socket, int) bool

//TransversalDirective provides a directive for transversing graphs
type TransversalDirective struct {
	Depth     int
	Order     TransversalOrder
	Revisits  VisitCaller
	Heuristic NodeOp
	AllNodes  bool
}

//MakeTransversalDirective creates a transversal directive
func MakeTransversalDirective(depth int, order TransversalOrder, visit VisitCaller, heuristic NodeOp) *TransversalDirective {
	if visit == nil {
		visit = defaultVisit
	}
	if heuristic == nil {
		heuristic = defaultHeuristic
	}

	return &TransversalDirective{
		Depth:     depth,
		Order:     order,
		Revisits:  visit,
		Heuristic: heumap(heuristic),
	}
}

//BFPreOrderDirective provides a copy of a breadth-first pre order rule
func BFPreOrderDirective(v VisitCaller, hx NodeOp) *TransversalDirective {
	return MakeTransversalDirective(-1, BFPreOrder, v, hx)
}

//BFPostOrderDirective provides a copy of a breath-first pre order rule
func BFPostOrderDirective(v VisitCaller, nx NodeOp) *TransversalDirective {
	return MakeTransversalDirective(-1, BFPostOrder, v, nx)
}

//DFPreOrderDirective provides a copy of a depth-first pre order rule
func DFPreOrderDirective(v VisitCaller, hx NodeOp) *TransversalDirective {
	return MakeTransversalDirective(-1, DFPreOrder, v, hx)
}

//NodeMaps represent the node map used by a iterator
type NodeMaps map[Nodes]bool

//NodeCaches represent the node cache used by a iterator
type NodeCaches []*NodeCache

//NodeCache provides a means of caching current node and current node iterator
type NodeCache struct {
	Node Nodes
	Itr  DeferIterator
}

//DFPostOrderDirective provides a copy of a depth-first pre order rule
func DFPostOrderDirective(v VisitCaller, hx NodeOp) *TransversalDirective {
	return MakeTransversalDirective(-1, DFPostOrder, v, hx)
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

		if t.directive.Depth != -1 {
			if t.directive.Depth <= t.WalkDepth() {
				return ErrBadIndex
			}
		}

		cur, err := cache.LastCache()

		if err != nil {
			t.current = nil
			return err
		}

		node, itr := cur.Node, cur.Itr

		if t.directive.Heuristic(node, t.keys[node]) != nil {
			cache.Uncache()
			return t.Next()
		}

		if t.directive.Revisits(node, t.visited.Valid(node)) {

			// if t.directive.Heuristic(node, t.keys[node]) != nil {
			// 	cache.Uncache()
			// 	return t.Next()
			// }

			t.visited.Add(node)
			t.current = node
			return nil
		}

		// if t.directive.Heuristic(node, t.keys[node]) != nil {
		// 	cache.Uncache()
		// 	return t.Next()
		// }

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

		if t.directive.Depth != -1 {
			if t.directive.Depth <= t.WalkDepth() {
				return ErrBadIndex
			}
		}

		curnode, err = cache.LastCache()

		if err != nil {
			t.current = nil
			return ErrBadIndex
		}

		cur := curnode.Itr
		node := curnode.Node

		if err := cur.Next(); err != nil {
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
			t.current = node
			return nil
		}

		curnode, ok := cur.Value().(*Socket)

		if !ok {
			t.current = nil
			return ErrBadEdgeType
		}

		co := curnode.To

		atomic.AddInt64(&t.walkdepth, 1)
		if !t.directive.Revisits(co, t.visited.Valid(co)) {
			return t.Next()
		}

		t.keys[co] = curnode

		if t.directive.Heuristic(co, t.keys[co]) != nil {
			return t.Next()
		}

		cache.AddCache(co)
		t.visited.Add(co)
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

		if t.directive.Depth != -1 {
			if t.directive.Depth <= t.WalkDepth() {
				return ErrBadIndex
			}
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

		if t.directive.Heuristic(no, t.keys[no]) != nil {
			// cache.Uncache()
			return t.Next()
		}

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

		if t.directive.Depth != -1 {
			if t.directive.Depth <= t.WalkDepth() {
				return ErrBadIndex
			}
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

		if t.directive.Heuristic(co, t.keys[co]) != nil {
			// cache.UncacheRight()
			// cache.Uncache()
			return t.Next()
		}

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

//Fx provides the default func type
type Fx func()

//Next provides a function type for the next function
type Next func() error

// Transversor provide functional transversal provider
type Transversor struct {
	directive     *TransversalDirective
	from, current Nodes
	keys          map[Nodes]*Socket
	depths        map[Nodes]int64
	started       int64
	walkdepth     int64
	visited       NodeMaps
	next          Next
	reset         Fx
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

//CreateGraphTransversor returns the correct transversal from the TransversalDirective else returns an error as second value
func CreateGraphTransversor(dir *TransversalDirective) (*Transversor, error) {

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

	return verso, nil
}

//DepthFirstPreOrderIterator returns a depth-first transverso
func DepthFirstPreOrderIterator(v VisitCaller, hx NodeOp) (*Transversor, error) {
	return CreateGraphTransversor(DFPreOrderDirective(v, hx))
}

//DepthFirstPostOrderIterator returns a depth-first transverso
func DepthFirstPostOrderIterator(v VisitCaller, hx NodeOp) (*Transversor, error) {
	return CreateGraphTransversor(DFPostOrderDirective(v, hx))
}

//BreadthFirstPreOrderIterator returns a depth-first transverso
func BreadthFirstPreOrderIterator(v VisitCaller, hx NodeOp) (*Transversor, error) {
	return CreateGraphTransversor(BFPreOrderDirective(v, hx))
}

//BreadthFirstPostOrderIterator returns a depth-first transverso
func BreadthFirstPostOrderIterator(v VisitCaller, hx NodeOp) (*Transversor, error) {
	return CreateGraphTransversor(BFPostOrderDirective(v, hx))
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

//GraphProc provides a means of running a function on all nodes of a graph
type GraphProc struct {
	trans *Transversor
	fx    NodeDop
	g     Graphs
}

//Search returns a GraphProc for depth-first
func Search(fx NodeDop, dir *TransversalDirective) (*GraphProc, error) {

	verso, err := CreateGraphTransversor(dir)

	if err != nil {
		return nil, err
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
	return p.fx(p.trans.Node(), p.trans.Key(), p.trans.WalkDepth())
}

//NodeDop is (Node Depth Operation) provide a type for running on graph iterators
type NodeDop func(Nodes, *Socket, int) error

//FilterNode provides a data type for inbetween filters
type FilterNode struct {
	Node   Nodes
	Socket *Socket
	Depth  int
}

//FilterBuilder provides a means of building a filter criteria
type FilterBuilder struct {
	filter *GraphFilter
	stack  NodeDop
}

//Filter returns a GraphFilter for transversing
func Filter(dir *TransversalDirective) (*FilterBuilder, error) {

	filter := &GraphFilter{}

	heu := dir.Heuristic

	proc, err := Search(func(n Nodes, soc *Socket, depth int) error {
		filter.paths = append(filter.paths, &FilterNode{n, soc, depth})
		return nil
	}, dir)

	if err != nil {
		return nil, err
	}

	dir.Heuristic = func(no Nodes, so *Socket) error {
		if err := heu(no, so); err != nil {
			return err
		}
		if filter.conditions == nil {
			return nil
		}
		if err := filter.conditions(no, so, proc.WalkDepth()); err != nil {
			return err
		}
		return nil
	}

	filter.proc = proc

	return &FilterBuilder{filter: filter}, nil
}

//Evaluator provides a means of adding a evaluation
func (f *FilterBuilder) Evaluator(eval NodeEval) *FilterBuilder {
	if f.filter == nil {
		return f
	}

	stack := f.stack

	f.stack = func(n Nodes, s *Socket, d int) error {
		if stack != nil {
			if err := stack(n, s, d); err != nil {
				return err
			}
		}
		if ok := eval(n, s, d); !ok {
			return ErrBadNode
		}
		return nil
	}

	return f
}

//Transverse returns the filter with set criteria
func (f *FilterBuilder) Transverse(n Nodes) *GraphFilter {
	fi := f.filter
	fi.conditions = f.stack
	f.filter = nil
	fi.Transverse(n)
	return fi
}

//GraphFilter provides a higher level filtering system ontop of the graphsearching framework
type GraphFilter struct {
	conditions NodeDop
	proc       *GraphProc
	paths      []*FilterNode
}

//reset resets the tranversal nodes
func (f *GraphFilter) reset() {
	f.paths = nil
	f.proc.Reset()
}

//Transverse sets the root node for tranversal
func (f *GraphFilter) Transverse(n Nodes) {
	f.reset()
	f.proc.Use(n)
}

//Next calls the filters tranversors next function
func (f *GraphFilter) Next() error {
	return f.proc.Next()
}

//Path returns the current path-ways found by the ops
func (f *GraphFilter) Path() []*FilterNode {
	return f.paths
}

//Nodes returns the current path-ways found by the ops
func (f *GraphFilter) Nodes() []Nodes {
	var nodes []Nodes
	paths := f.Path()

	for _, f := range paths {
		nodes = append(nodes, f.Node)
	}

	paths = nil
	return nodes
}

// GraphSearcher defines interface rules for searches
type GraphSearcher interface {
	FindOne() (Nodes, error)
	FindAll() ([]Nodes, error)
}

// EvaluateNode provides a function type for the linear searching algorithm
type EvaluateNode func(Nodes) bool

// LinearGraphSearch provides a simple,top-down search system for a graph and returns the result when a match is found
type LinearGraphSearch struct {
	graph Graphs
	ro    sync.Mutex
}

// NewLinearGraphSearch returns a new LinearGraphSearch
func NewLinearGraphSearch(g Graphs) *LinearGraphSearch {
	ls := LinearGraphSearch{
		graph: g,
	}
	return &ls
}

// FindOne runs through and retuns the first matching result or an error
func (nl *LinearGraphSearch) FindOne(ev EvaluateNode) (Nodes, error) {
	nl.ro.Lock()
	defer nl.ro.Unlock()

	var res Nodes

	ns := nl.graph.nodeSet()

	ns.Each(func(n Nodes, _ int, stop func()) {
		if ev(n) {
			res = n
			stop()
		}
	})

	if res == nil {
		return nil, ErrNotFound
	}

	return res, nil
}

// FindAll runs through and retuns all eatching results or an error
func (nl *LinearGraphSearch) FindAll(ev EvaluateNode) ([]Nodes, error) {
	nl.ro.Lock()
	defer nl.ro.Unlock()

	var res []Nodes

	ns := nl.graph.nodeSet()

	ns.Each(func(n Nodes, _ int, stop func()) {
		if ev(n) {
			res = append(res, n)
		}
	})

	if res == nil {
		return nil, ErrNotFound
	}

	return res, nil
}
