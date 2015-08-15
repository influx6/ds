package ds

import (
	"fmt"
	"sync/atomic"

	"github.com/influx6/sequence"
)

//DFSDax provides a dax wrapper over a DFSIterator
func DFSDax(dfs *DFSIterator) *daxdfs {
	return &daxdfs{dfs}
}

//BFSDax provides a dax wrapper over a BFSIterator
func BFSDax(bfs *BFSIterator) *daxbfs {
	return &daxbfs{bfs}
}

//NewNodeCache returns a new NodeCache
func NewNodeCache(n Nodes) *NodeCache {
	return &NodeCache{
		Node: n,
		Itr:  n.Arcs(),
	}
}

//String returns the string of the node
func (n *NodeCache) String() string {
	return fmt.Sprintf("%+s", n.Node.Value())
}

//NewDFSIterator returns a depth-first traversal interator
func NewDFSIterator(g Graphs, depth int) *DFSIterator {
	return &DFSIterator{
		GraphIterator: BaseGraphIterator(g, depth),
	}
}

//NewBFSIterator returns a breadth-first traversal interator
func NewBFSIterator(g Graphs, depth int) *BFSIterator {
	return &BFSIterator{
		GraphIterator: BaseGraphIterator(g, depth),
	}
}

//Clone creates another iterator from the graph
func (d *DFSIterator) Clone() sequence.Iterable {
	return NewDFSIterator(d.graph, d.depth)
}

//Clone creates another iterator from the graph
func (d *BFSIterator) Clone() sequence.Iterable {
	return NewBFSIterator(d.graph, d.depth)
}

//Value returns the current node
func (d *DFSIterator) Value() interface{} {
	return d.current
}

//Value returns the current node
func (d *BFSIterator) Value() interface{} {
	return d.current
}

//hasNext returns true/false if it has more elements
func (d *DFSIterator) hasNext() bool {

	if len(d.cache) <= 0 && d.started > 0 {
		return false
	}

	if len(d.cache) <= 0 {
		first, err := d.graph.nodeSet().FirstNode()

		if err != nil {
			return false
		}

		atomic.StoreInt64(&d.started, 1)
		d.addCache(first)
	}

	return true
}

//Next moves to the next element
func (d *DFSIterator) Next() error {
	if !d.hasNext() {
		return sequence.ErrBADINDEX
	}

	cur, err := d.lastCache()

	if err != nil {
		d.current = nil
		return err
	}

	node, itr := cur.Node, cur.Itr

	if !d.visited[node] {
		d.visited[node] = true
		d.current = node
		return nil
	}

	err = itr.Next()

	if err != nil {
		d.unCache()
		return d.Next()
	}

	cursoc, ok := itr.Value().(*Socket)

	if !ok {
		d.current = nil
		return ErrBadEdgeType
	}

	node = cursoc.To

	// log.Printf("Checking next visited: %+s %+s", node.Value(), d.visited[node])
	if d.visited[node] {
		return d.Next()
	}

	d.addCache(node)

	return d.Next()
}

//hasNext returns true/false if it has more elements
func (d *BFSIterator) hasNext() bool {
	if len(d.cache) <= 0 && d.started > 0 {
		return false
	}

	if len(d.cache) <= 0 {
		first, err := d.graph.nodeSet().FirstNode()

		if err != nil {
			return false
		}

		atomic.StoreInt64(&d.started, 1)
		d.addCache(first)
	}

	return true
}

//Next moves to the next element
func (d *BFSIterator) Next() error {
	if !d.hasNext() {
		return sequence.ErrBADINDEX
	}

	cur, err := d.lastCache()

	if err != nil {
		d.current = nil
		return err
	}

	node, itr := cur.Node, cur.Itr

	if !d.visited[node] {
		d.visited[node] = true
		d.current = node
		return nil
	}

	d.unCache()

	for itr.Next() == nil {
		sock, ok := itr.Value().(*Socket)
		if !ok {
			continue
		}
		d.addCache(sock.To)
	}

	return d.Next()
}

//DSF returns a GraphProc for depth-first
func DSF(fx GraphHandlers, g Graphs, depth int) *GraphProc {
	return &GraphProc{
		GraphIterable: DFSDax(NewDFSIterator(g, depth)),
		fx:            fx,
	}
}

//BSF returns a GraphProc for depth-first
func BSF(fx GraphHandlers, g Graphs, depth int) *GraphProc {
	return &GraphProc{
		GraphIterable: BFSDax(NewBFSIterator(g, depth)),
		fx:            fx,
	}
}
