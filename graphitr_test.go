package ds

import "testing"

func TestDFS(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7, 8)
	gs.Bind(1, 3, 0)
	gs.Bind(3, 4, 0)
	gs.Bind(8, 3, 0)
	gs.Bind(4, 8, 0)
	gs.Bind(1, 7, 0)

	dfs := DSF(func(n Nodes) error {
		return nil
	}, gs, -1)

	for dfs.Next() == nil {
	}

	if len(dfs.Unvisited()) > 2 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	dfs.Reset()
}

func TestBFS(t *testing.T) {
	var gs = NewGraph()
	gs.Add(1, 3, 4, 5, 6, 7)
	gs.Bind(1, 3, 0)
	gs.Bind(5, 4, 0)
	gs.Bind(1, 7, 0)

	bfs := BSF(func(n Nodes) error {
		return nil
	}, gs, -1)

	for bfs.Next() == nil {
	}

	if len(bfs.Unvisited()) > 3 {
		t.Fatal("Incorrect number of unvisited nodes")
	}

	bfs.Reset()
}
