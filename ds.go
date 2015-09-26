package ds

import "errors"

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

	//Unvisited provides a function map for unvisited nodes
	Unvisited func(Graphs) []Nodes

	//GNode returns a node type
	GNode func() Nodes

	//UseNode provides the use node type
	UseNode func(Nodes)
)
