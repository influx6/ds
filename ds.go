package ds

import (
	"errors"
	"sync"

	"github.com/influx6/flux"
	"github.com/influx6/sequence"
)

var (
	//ErrBadList indicates this list is invalid
	ErrBadList = errors.New("Invalid List")
	//ErrBadNode indicates this node does not belong
	ErrBadNode = errors.New("BadNode not in list")
	//ErrNoEdge indicates this node does not belong in the list of edged
	ErrNoEdge = errors.New("Node not in edges")
)

type (

	//Equalers define an interface with Equal func
	Equalers interface {
		// Values
		Equals(interface{}) bool
	}

	//EqualSet defines a set of
	set []Equalers

	//BaseSet provides an implementation for different set types
	baseset struct {
		set set
		rw  *sync.RWMutex
	}

	//NodeSet provides a set implementation for graph nodes
	NodeSet struct {
		set   *baseset
		dirty int64
	}

	//DeferNodeSet defines a set implementation for differered nodes
	DeferNodeSet struct {
		set   *baseset
		dirty int64
	}

	//Values defines an interface with Values
	Values interface {
		Value() interface{}
	}

	//NodePush defines a function for setting node links
	NodePush func(*DeferNode) *DeferNode

	//DeferIterator provides and defines methods for defer iteratore
	DeferIterator interface {
		sequence.Iterable
		Reset2Tail()
		Reset2Root()
	}

	//MovableDeferIterator defines iterators that can move both forward and backward
	MovableDeferIterator interface {
		DeferIterator
		Previous() error
		HasPrevious() bool
	}

	//DeferNode represents a standard node meeting DeferNode requirements
	DeferNode struct {
		data interface{}
		list *DeferList
		next NodePush
		prev NodePush
	}

	//DeferListIterator provides an iterator for DeferredList
	DeferListIterator struct {
		list    *DeferList
		current *DeferNode
		state   int64
	}

	//DeferList represents a sets of linkedlist meeting DeferList requirements
	DeferList struct {
		tail NodePush
		root NodePush
		size int64
	}

	//Graphs represents the standard Graph interface
	Graphs interface {
		// sequence.SizableSequencable
		Contains(interface{}) bool
		Get(interface{}) Nodes
		Add(...interface{})
		AddNode(Nodes)
		AddForeignNode(r Nodes)
		Bind(interface{}, interface{}, int) (*Socket, bool)
		UnBind(interface{}, interface{}) bool
		IsBound(interface{}, interface{}) bool
		// UnBindAll(interface{}, interface{}) bool
	}

	//Nodes represents the standard Graph interface
	Nodes interface {
		flux.StringMappable
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
	}

	//Graph represent a standard structure of nodes
	Graph struct {
		nodes *NodeSet
	}

	//Node represents an element in the graph
	Node struct {
		flux.StringMappable
		data  interface{}
		arcs  *DeferList
		graph Graphs
	}

	//Socket represents a connection between two nodes
	Socket struct {
		flux.StringMappable
		To     Nodes
		From   Nodes
		Weight int
	}
)
