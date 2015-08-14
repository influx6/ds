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

	for itr.HasNext() {
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
		return dx.Value() == n.data
	}

	// dc, ok := d.(*Node)
	//
	// if ok {
	// 	return dc.data == n.data
	// }
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

	for itr.HasNext() {
		sock, _ := itr.Value().(*Socket)

		if sock.To != r {
			itr.Next()
			continue
		}

		return sock, nil
	}

	return nil, ErrNoEdge
}

//HasEdge returns if a node is connected to this node
func (n *Node) HasEdge(r Nodes) bool {
	itr := n.arcs.Iterator()

	for itr.HasNext() {
		sock, _ := itr.Value().(*Socket)

		if sock.To == r {
			itr.Next()
			return true
		}
		itr.Next()
	}

	return false
}

//RemovalEdges removes all edges of this node
func (n *Node) RemovalEdges() {
	itr := n.arcs.Iterator()

	for itr.HasNext() {
		sock, _ := itr.Value().(*Socket)
		sock.Close()
		itr.Next()
	}

	n.arcs.Clear()
}

func (n *Node) disconnect(r Nodes, one bool) {
	itr := n.arcs.Iterator()

	for itr.HasNext() {
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

			itr.Next()
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

//Bind binds the two nodes of this values
func (n *Graph) Bind(r, f interface{}, we int) (*Socket, bool) {
	rx, rk := n.nodes.GetNode(r)
	fx, fk := n.nodes.GetNode(f)

	if !rk || !fk {
		return nil, false
	}

	return rx.Connect(fx, we), true
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
