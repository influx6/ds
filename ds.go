package ds

import (
	"errors"
	"sync"

	"github.com/influx6/flux"
	"github.com/influx6/sequence"
)

var (
	//ErrBadIterator indicates the iterator has an issue
	ErrBadIterator = errors.New("Invalid Iterator State")
	//ErrBadList indicates this list is invalid
	ErrBadList = errors.New("Invalid List")
	//ErrBadIndex indicates this node index does not belong
	ErrBadIndex = errors.New("BadIndex value")
	//ErrBadNode indicates this node does not belong
	ErrBadNode = errors.New("BadNode not in list")
	//ErrEmpty to indicate empty list or graph
	ErrEmpty = errors.New("A list is empty")
	//ErrNoEdge indicates this node does not belong in the list of edged
	ErrNoEdge = errors.New("Node not in edges")
	//ErrBadBind indicates this node does not belong
	ErrBadBind = errors.New("BadBind unable to bind nodes")
	//ErrBadEdgeType indicates that the value of a iterator is not a *Socket
	ErrBadEdgeType = errors.New("value is not a *Socket type")
)

type (

	//Equalers define an interface with Equal func
	Equalers interface {
		// Values
		Equals(interface{}) bool
		// Equal(interface{}) bool
	}

	//EqualSet defines a set of
	set []Equalers

	//BaseSet provides an implementation for different set types
	baseset struct {
		set set
		rw  *sync.RWMutex
	}

	//String provides a super-type alias for strings
	String string

	//StringSet provides a set impl for strings
	StringSet struct {
		set   *baseset
		dirty int64
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
		Arcs() DeferIterator
		String() string
	}

	//Graph represent a standard structure of nodes
	Graph struct {
		nodes *NodeSet
		uid   string
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
		Attrs  *StringSet
		To     Nodes
		From   Nodes
		Weight int
	}

	//NodeMaps represent the node map used by a iterator
	NodeMaps map[Nodes]bool

	//NodeCaches represent the node cache used by a iterator
	NodeCaches []*NodeCache

	//NodeCache provides a means of caching current node and current node iterator
	NodeCache struct {
		Node Nodes
		Itr  DeferIterator
	}

	//TransversalOrder provides a order for TranversalDirective
	TransversalOrder string

	//VisitCaller provides a type for visit checks
	VisitCaller func(Nodes, bool) bool

	//TransversalDirective provides a directive for transversing graphs
	TransversalDirective struct {
		Depth     int
		Order     TransversalOrder
		Revisits  VisitCaller
		Heuristic NodeOp
		AllNodes  bool
	}

	//Unvisited provides a function map for unvisited nodes
	Unvisited func(Graphs) []Nodes

	//Next provides a function type for the next function
	Next func() error

	//GNode returns a node type
	GNode func() Nodes

	//Fx provides the default func type
	Fx func()

	//UseNode provides the use node type
	UseNode func(Nodes)

	//Transversor provide functional transversal provider
	Transversor struct {
		directive     *TransversalDirective
		from, current Nodes
		keys          map[Nodes]*Socket
		started       int64
		walkdepth     int64
		visited       NodeMaps
		next          Next
		reset         Fx
	}

	//NodeOp provide a type for running on graph iterators
	NodeOp func(Nodes, *Socket) error

	//NodeDop is (Node Depth Operation) provide a type for running on graph iterators
	NodeDop func(Nodes, *Socket, int) error

	//NodeEval provides a evalutor format type
	NodeEval func(Nodes, *Socket, int) bool

	//GraphProc provides a means of running a function on all nodes of a graph
	GraphProc struct {
		trans *Transversor
		fx    NodeDop
		g     Graphs
	}

	//GraphFilter provides a higher level filtering system ontop of the graphsearching framework
	GraphFilter struct {
		conditions NodeDop
		proc       *GraphProc
		paths      []*FilterNode
	}

	//FilterNode provides a data type for inbetween filters
	FilterNode struct {
		Node   Nodes
		Socket *Socket
		Depth  int
	}

	//FilterBuilder provides a means of building a filter criteria
	FilterBuilder struct {
		filter *GraphFilter
		stack  NodeDop
	}
)

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
