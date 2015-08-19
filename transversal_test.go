package ds

import "testing"

func TestDFSPreHeuristic(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)

	dfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, DFPreOrderDirective(nil, func(n Nodes, _ *Socket) error {
		if n.Value() == 4 {
			return ErrBadNode
		}
		return nil
	}))

	if err != nil {
		t.Fatal(err)
	}

	dfs.Use(gs.Get(1))

	for dfs.Next() == nil {
	}

	unvisited := len(dfs.Unvisited())
	if unvisited > 4 {
		t.Fatalf("Incorrect number of unvisited nodes expected 2 got %d", unvisited)
	}

	dfs.Reset()
}
func TestDFSPre(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)

	dfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, DFPreOrderDirective(nil, nil))

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

func TestBFSPreHeuristic(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(4, 6, 0)
	gs.Bind(5, 4, 0)
	gs.Bind(1, 7, 0)
	gs.Bind(1, 6, 0)
	gs.Bind(6, 5, 0)

	bfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, BFPreOrderDirective(nil, func(node Nodes, _ *Socket) error {
		if node.Value() == 6 {
			return ErrBadNode
		}
		return nil
	}))

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

	bfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, BFPreOrderDirective(nil, nil))

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

	bfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, BFPostOrderDirective(nil, nil))

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

	dfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, DFPostOrderDirective(nil, nil))

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

func TestDFSPostWithHeurisitic(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)
	gs.Bind(7, 6, 0)

	dfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, DFPostOrderDirective(nil, func(node Nodes, _ *Socket) error {
		if node.Value() == 3 {
			return ErrBadNode
		}
		return nil
	}))

	if err != nil {
		t.Fatal(err)
	}

	dfs.Use(gs.Get(1))

	for dfs.Next() == nil {
	}

	if len(dfs.Unvisited()) > 4 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	dfs.Reset()
}

func TestBFSPostHeuristic(t *testing.T) {
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

	bfs, err := Search(func(n Nodes, sock *Socket, _ int) error {
		if n == nil {
			t.Fatal("Received a nil node")
		}
		return nil
	}, BFPostOrderDirective(nil, func(node Nodes, _ *Socket) error {
		if node.Value() == 4 {
			return ErrBadNode
		}
		return nil
	}))

	if err != nil {
		t.Fatal(err)
	}

	bfs.Use(gs.Get(1))

	for bfs.Next() == nil {
	}

	if vl := len(bfs.Unvisited()); vl > 3 {
		t.Fatal("Incorrect number of unvisited nodes:", vl)
	}

	bfs.Reset()
}
