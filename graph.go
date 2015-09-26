package ds

import (
	"fmt"
	"strings"

	"code.google.com/p/go-uuid/uuid"

	"github.com/influx6/flux"
)

//Values defines an interface with Values
type Values interface {
	Value() interface{}
}

//Nodes represents the standard Graph interface
type Nodes interface {
	Values
	Equalers
	//Represents the archs/edges of this nodes
	Sockets() *DeferList
	//Bind the supplied node to this node
	Connect(Nodes, int) *Socket
	//Unbind the supplied node from this one
	GetEdge(Nodes) (*Socket, error)
	HasEdge(Nodes) bool
	RemovalEdges()
	Disconnect(Nodes)
	DisconnectOne(Nodes)
	//ChangeGraph changes the underline graph this nodes belongs to
	ChangeGraph(Graphs)
	//Graph returns the graph of the node
	Graph() Graphs
	Arcs() DeferIterator
	String() string
}

//Socket represents a connection between two nodes
type Socket struct {
	flux.Collector
	Attrs  *StringSet
	To     Nodes
	From   Nodes
	Weight int
}

//NewSocket creates a new socket between two nodes with a set weight
func NewSocket(from, to Nodes, weight int) *Socket {
	return &Socket{
		Collector: flux.NewCollector(),
		Attrs:     NewStringSet(),
		To:        to,
		From:      from,
		Weight:    weight,
	}
}

//Close closes and destroys the socket between the two nodes
func (s *Socket) Close() {
	s.To.Disconnect(s.From)
	s.From = nil
	s.To = nil
	s.Clear()
}

//Node represents an element in the graph
type Node struct {
	data  interface{}
	arcs  *DeferList
	graph Graphs
}

//ChangeValue changes the value of the node
func (n *Node) ChangeValue(d interface{}) {
	n.data = d
}

//ChangeGraph changes the graph this nodes is attached to
func (n *Node) ChangeGraph(g Graphs) {
	if n.graph == g {
		return
	}

	n.graph = g
	itr := n.arcs.Iterator()

	for itr.Next() == nil {
		cur, ok := itr.Value().(*Socket)
		if ok {
			cur.Close()
		}
	}

	n.arcs.Clear()
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
		data:  d,
		arcs:  List(),
		graph: g,
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

//Graphs represents the standard Graph interface
type Graphs interface {
	// sequence.SizableSequencable
	UID() string
	Contains(interface{}) bool
	Get(interface{}) Nodes
	Add(...interface{})
	AddNode(Nodes)
	AddForeignNode(r Nodes)
	Bind(interface{}, interface{}, int) (*Socket, bool)
	UnBind(interface{}, interface{}) bool
	BindNodes(Nodes, Nodes, int) (*Socket, bool)
	UnBindNodes(Nodes, Nodes) bool
	IsBound(interface{}, interface{}) bool
	nodeSet() *NodeSet
	Length() int
	// UnBindAll(interface{}, interface{}) bool
}

//Graph represent a standard structure of nodes
type Graph struct {
	nodes *NodeSet
	uid   string
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

//Length returns the total size of nodes in this graph
// func (n *Graph) Length() int {
// 	return n.nodes.Length()
// }

//UID returns the auto generated uuid for this graph
func (n *Graph) UID() string {
	return n.uid
}

func (n *Graph) String() string {
	return strings.Join([]string{
		fmt.Sprintf("<Graph UID='%s'>\n", n.uid),
		"-> Nodes Length",
		fmt.Sprintf("%d", n.Length()),
		"</Graph>",
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

//UnvisitedUtil returns the current set of unvisited nodes
func UnvisitedUtil(g Graphs, visited NodeMaps) []Nodes {
	unvs := []Nodes{}

	if len(visited) <= 0 {
		return g.nodeSet().AllNodes()
	}

	g.nodeSet().EachNode(func(nx Nodes) {
		if !visited[nx] {
			unvs = append(unvs, nx)
		}
	})

	return unvs
}

//Reset resets the cache
func (d *NodeMaps) Reset() {
	*d = make(map[Nodes]bool)
}

//Add removes the last node cache
func (d *NodeMaps) Add(n Nodes) {
	(*d)[n] = true
}

//Valid removes the last node cache
func (d *NodeMaps) Valid(n Nodes) bool {
	return (*d)[n]
}

//Length returns the length node cache
func (d *NodeMaps) Length() int {
	return len(*d)
}

//Length returns the length the cache
func (d *NodeCaches) Length() int {
	return len(*d)
}

//Reset resets the cache
func (d *NodeCaches) Reset() {
	*d = (*d)[:0]
}

//UncacheRight removes the first node cache and closes the cache object
func (d *NodeCaches) UncacheRight() {
	nx := (*d)[0]
	*d = (*d)[1:]
	nx.Close()
}

//Uncache removes the last node cache and closes the cache object
func (d *NodeCaches) Uncache() {
	nx := (*d)[len(*d)-1]
	*d = (*d)[:len(*d)-1]
	nx.Close()
}

//NthCache let you get a item at a point in the cache
func (d *NodeCaches) NthCache(ind int) (*NodeCache, error) {
	if ind >= d.Length() {
		return nil, ErrBadIndex
	}

	var loc int

	if ind < 0 {
		loc = d.Length() - ind
	} else {
		loc = ind
	}

	return (*d)[loc], nil
}

//FirstCache returns the last node cached
func (d *NodeCaches) FirstCache() (*NodeCache, error) {
	clen := len(*d)

	if clen <= 0 {
		return nil, ErrBadIndex
	}

	return (*d)[0], nil
}

//LastCache returns the last node cached
func (d *NodeCaches) LastCache() (*NodeCache, error) {
	clen := len(*d)

	if clen <= 0 {
		return nil, ErrBadIndex
	}

	return (*d)[clen-1], nil
}

//AddCache adds a node to the cache
func (d *NodeCaches) AddCache(n Nodes) {
	*d = append(*d, NewNodeCache(n))
}

//VisitMaps returns a new NodeCache
func VisitMaps() NodeMaps {
	return make(NodeMaps)
}

//NewCache returns a new NodeCache
func NewCache() NodeCaches {
	return make(NodeCaches, 0)
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
	if n.Node != nil {
		return fmt.Sprintf("%+s", n.Node.Value())
	}
	return ""
}

//Close destroys this cache
func (n *NodeCache) Close() error {
	n.Node = nil
	n.Itr = nil
	return nil
}
