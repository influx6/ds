package ds

import (
	"log"
	"testing"
)

func TestDFSPre(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)

	dfs, err := Search(func(n Nodes, sock *Socket) error {
		log.Println("DFS:", n, sock)
		return nil
	}, DFPreOrderDirective(-1, nil))

	if err != nil {
		t.Fatal(err)
	}

	dfs.Use(gs.Get(1))

	for dfs.Next() == nil {
	}

	unvisited := len(dfs.Unvisited())
	if unvisited > 2 {
		t.Fatalf("Incorrect number of unvisited nodes expected 2 got %d", unvisited)
	}

	dfs.Reset()
}

func TestBFSPre(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(4, 6, 0)
	gs.Bind(5, 4, 0)
	gs.Bind(1, 7, 0)
	gs.Bind(1, 6, 0)
	gs.Bind(6, 5, 0)

	bfs, err := Search(func(n Nodes, sock *Socket) error {
		log.Println("BFS:", n, sock)
		return nil
	}, BFPreOrderDirective(-1, nil))

	bfs.Use(gs.Get(1))
	if err != nil {
		t.Fatal(err)
	}

	for bfs.Next() == nil {
	}

	if len(bfs.Unvisited()) > 3 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	bfs.Reset()
}

func TestBFSPost(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(4, 6, 0)
	gs.Bind(4, 5, 0)
	gs.Bind(5, 4, 0)
	gs.Bind(1, 7, 0)
	gs.Bind(1, 6, 0)
	gs.Bind(6, 3, 0)

	bfs, err := Search(func(n Nodes, sock *Socket) error {
		log.Println("BFS-Post:", n, sock)
		return nil
	}, BFPostOrderDirective(-1, nil))

	if err != nil {
		t.Fatal(err)
	}

	bfs.Use(gs.Get(1))

	for bfs.Next() == nil {
	}

	if len(bfs.Unvisited()) > 3 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	bfs.Reset()
}

func TestDFSPost(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)
	gs.Bind(7, 6, 0)

	dfs, err := Search(func(n Nodes, sock *Socket) error {
		log.Println("DFS-Post:", n, sock)
		return nil
	}, DFPostOrderDirective(-1, nil))

	if err != nil {
		t.Fatal(err)
	}

	dfs.Use(gs.Get(1))

	for dfs.Next() == nil {
	}

	if len(dfs.Unvisited()) > 2 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	dfs.Reset()
}
