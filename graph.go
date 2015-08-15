package ds

import (
	"fmt"
	"strings"

	"code.google.com/p/go-uuid/uuid"

	"github.com/influx6/flux"
)

//NewSocket creates a new socket between two nodes with a set weight
func NewSocket(from, to Nodes, weight int) *Socket {
	return &Socket{
		StringMappable: flux.NewCollector(),
		To:             to,
		From:           from,
		Weight:         weight,
	}
}

//Close closes and destroys the socket between the two nodes
func (s *Socket) Close() {
	s.To.Disconnect(s.From)
	s.From = nil
	s.To = nil
	s.Clear()
}

//ChangeValue changes the value of the node
func (n *Node) ChangeValue(d interface{}) {
	n.data = d
}

//ChangeGraph changes the graph this nodes is attached to
func (n *Node) ChangeGraph(g Graphs) {
	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		cur, ok := itr.Value().(*Socket)
		if ok {
			cur.Close()
		}
	}

	n.arcs.Clear()
	n.graph = g
}

//Graph returns the graph we are connected with
func (n *Node) Graph() Graphs {
	return n.graph
}

//Equals returns bool if the value is equal
func (n *Node) Equals(d interface{}) bool {
	dx, ok := d.(Nodes)

	if ok {
		if dx == n {
			return true
		}
		return dx.Value() == n.data
	}

	return n.data == d
}

//Sockets returns the list of arcs/sockets
func (n *Node) Sockets() *DeferList {
	return n.arcs
}

//DisconnectOne removes all edges of the giving node
func (n *Node) DisconnectOne(r Nodes) {
	n.disconnect(r, true)
}

//Disconnect removes all edges of the giving node
func (n *Node) Disconnect(r Nodes) {
	n.disconnect(r, false)
}

//GetEdge returns the socket that connects a node is connected to this node
func (n *Node) GetEdge(r Nodes) (*Socket, error) {
	if r == nil {
		return nil, ErrNoEdge
	}

	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		sock, _ := itr.Value().(*Socket)

		if sock.To != r {
			continue
		}

		return sock, nil
	}

	return nil, ErrNoEdge
}

//HasEdge returns if a node is connected to this node
func (n *Node) HasEdge(r Nodes) bool {
	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		sock, _ := itr.Value().(*Socket)

		if sock.To == r {
			return true
		}
	}

	return false
}

//RemovalEdges removes all edges of this node
func (n *Node) RemovalEdges() {
	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		sock, _ := itr.Value().(*Socket)
		sock.Close()
	}

	n.arcs.Clear()
}

func (n *Node) disconnect(r Nodes, one bool) {
	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		sock, _ := itr.Value().(*Socket)

		if sock.To == r {
			key, ok := itr.Key().(*DeferNode)

			if ok {
				key.Reset()
			}

			sock.Close()

			if one {
				break
			}

		}
	}
}

//Connect the provide node to itself
func (n *Node) Connect(r Nodes, weight int) *Socket {
	var socket *Socket
	var err error

	socket, err = n.GetEdge(r)

	if err == nil {
		return socket
	}

	socket = NewSocket(n, r, weight)
	// _ = r.Connect(n, weight)
	n.arcs.AppendElement(socket)
	return socket
}

//NewGraphNode returns a new graph node
func NewGraphNode(d interface{}, g Graphs) *Node {
	return &Node{
		StringMappable: flux.NewCollector(),
		data:           d,
		arcs:           List(),
		graph:          g,
	}
}

//String returns the string value
func (n *Node) String() string {
	return fmt.Sprintf("%+v", n.Value())
}

//Arcs returns an iterator of all the arcs/edges of this node
func (n *Node) Arcs() DeferIterator {
	itr := n.arcs.Iterator()
	return itr.(DeferIterator)
}

//Value returns the value of the node
func (n *Node) Value() interface{} {
	return n.data
}

//Length returns the size of the graph
func (n *Graph) Length() int {
	return n.nodes.Length()
}

//Get as a node into the graph
func (n *Graph) Get(r interface{}) Nodes {
	nz, _ := n.nodes.GetNode(r)
	return nz
}

//Contains returns true wether the graph has the element
func (n *Graph) Contains(r interface{}) bool {
	_, f := n.nodes.GetNode(r)
	return f
}

//AddNode as a node into the graph and sets the node graph to this graph,thereby clearing all previos connection
func (n *Graph) AddNode(r Nodes) {
	r.ChangeGraph(n)
	n.nodes.AddNode(r)
}

//AddForeignNode as a node into the graph without setting the node graph to this graph,thereby clearing all previos connection but note if this nodes value is the same with another node in this,this will be rejected
func (n *Graph) AddForeignNode(r Nodes) {
	if !n.Contains(r.Value()) {
		// r.ChangeGraph(n)
		n.nodes.AddNode(r)
	}
}

//UID returns the auto generated uuid for this graph
func (n *Graph) UID() string {
	return n.uid
}

func (n *Graph) String() string {
	return strings.Join([]string{
		fmt.Sprintf("<Graph UID='%s'>\n", n.uid),
		"-> Nodes Length",
		fmt.Sprintf("%d", n.Length()),
		"\n",
		"-> Content (DeptFirst Mode):",
		"\n</Graph>",
	}, " ")
}

//Add as a new node into the graph
func (n *Graph) Add(r ...interface{}) {
	for _, v := range r {
		n.nodes.AddNode(NewGraphNode(v, n))
	}
}

//BindNodes binds the two nodes of this values
func (n *Graph) BindNodes(r, f Nodes, we int) (*Socket, bool) {
	if r.Graph() != n || f.Graph() != n {
		return nil, false
	}

	return r.Connect(f, we), true
}

//UnBindNodes binds the two nodes of this values
func (n *Graph) UnBindNodes(r, f Nodes) bool {
	if r.Graph() != n || f.Graph() != n {
		return false
	}

	r.Disconnect(f)
	return true
}

//Bind binds the two nodes of this values
func (n *Graph) Bind(r, f interface{}, we int) (*Socket, bool) {
	rx, rk := n.nodes.GetNode(r)
	fx, fk := n.nodes.GetNode(f)

	if !rk || !fk {
		return nil, false
	}

	return rx.Connect(fx, we), true
}

//nodeSet returns the internal graph nodeset
func (n *Graph) nodeSet() *NodeSet {
	return n.nodes
}

//UnBind unbinds the two nodes of this values
func (n *Graph) UnBind(r, f interface{}) bool {
	rx, rk := n.nodes.GetNode(r)
	fx, fk := n.nodes.GetNode(f)

	if !rk || !fk {
		return false
	}

	rx.Disconnect(fx)
	return true
}

//IsBound returns true if both nodes are bound
func (n *Graph) IsBound(r, f interface{}) bool {
	rx, rk := n.nodes.GetNode(r)
	fx, fk := n.nodes.GetNode(f)

	if !rk || !fk {
		return false
	}

	return rx.HasEdge(fx)
}

//NewGraph returns a new graph instance
func NewGraph() *Graph {
	return &Graph{
		nodes: NewNodeSet(),
		uid:   uuid.New(),
	}
}

//BaseGraphIterator returns a baselevel iterator struct
func BaseGraphIterator(g Graphs, depth int) *GraphIterator {
	return &GraphIterator{
		graph:   g,
		depth:   depth,
		visited: make(map[Nodes]bool),
	}
}

//lastCache returns the last node cached
func (d *GraphIterator) lastCache() (*NodeCache, error) {
	clen := len(d.cache)

	if clen <= 0 {
		return nil, ErrBadIndex
	}

	return d.cache[clen-1], nil
}

func (d *GraphIterator) addCache(n Nodes) {
	d.cache = append(d.cache, NewNodeCache(n))
}

//Length returns the length of the graph
func (d *GraphIterator) Length() int {
	return d.graph.nodeSet().Length()
}

//Key returns the graph itself
func (d *GraphIterator) Key() interface{} {
	return d.graph
}

//Unvisited returns the current set of unvisited nodes before a .Reset()
//called after a search if Reset() has been called a empty list is returned
func (d *GraphIterator) Unvisited() []Nodes {
	unvs := []Nodes{}

	if len(d.visited) <= 0 {
		return d.graph.nodeSet().AllNodes()
	}

	d.graph.nodeSet().EachNode(func(nx Nodes) {
		if !d.visited[nx] {
			unvs = append(unvs, nx)
		}
	})
	return unvs
}

//Reset resets the iterator
func (d *GraphIterator) Reset() {
	d.cache = d.cache[:0]
	d.visited = make(map[Nodes]bool)
	d.current = nil
}

func (d *GraphIterator) unCache() {
	d.cache = d.cache[:len(d.cache)-1]
}

//Next calls the iterator next call
func (d *GraphProc) Next() error {
	err := d.GraphIterable.Next()

	if err != nil {
		return err
	}

	return d.fx(d.Node())
}

//Next calls the iterator next call
func (d *daxdfs) Next() error {
	return d.itr.Next()
}

//Reset resets the iterator
func (d *daxdfs) Reset() {
	d.itr.Reset()
}

//Unvisited returns the unvisited nodes
func (d *daxdfs) Unvisited() []Nodes {
	return d.itr.Unvisited()
}

//Node returns the current node
func (d *daxdfs) Node() Nodes {
	return d.itr.Value().(Nodes)
}

//Unvisited returns the unvisited nodes
func (d *daxbfs) Unvisited() []Nodes {
	return d.itr.Unvisited()
}

//Next calls the iterator next call
func (d *daxbfs) Next() error {
	return d.itr.Next()
}

//Reset resets the iterator
func (d *daxbfs) Reset() {
	d.itr.Reset()
}

//Node returns the current node
func (d *daxbfs) Node() Nodes {
	return d.itr.Value().(Nodes)
}
