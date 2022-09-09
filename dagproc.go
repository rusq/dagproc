package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yourbasic/graph"
)

func main() {
	var Nodes = []Node{
		{ID: "0", Delay: 1, ParentIDs: []string{}},
		{ID: "1", Delay: 1, ParentIDs: []string{"0"}},
		{ID: "2", Delay: 1, ParentIDs: []string{"1"}},
		{ID: "3", Delay: 4, ParentIDs: []string{"1"}},
		{ID: "4", Delay: 1, ParentIDs: []string{"2"}},
		{ID: "5", Delay: 1, ParentIDs: []string{"4"}},
		{ID: "6", Delay: 2, ParentIDs: []string{"2", "3"}},
		{ID: "7", Delay: 1, ParentIDs: []string{"1", "3", "8"}},
		{ID: "8", Delay: 1, ParentIDs: []string{"5", "6"}},
	}
	g := NewGraph(Nodes)

	order, ok := graph.TopSort(g.g)
	if !ok {
		log.Fatal("poop")
	}

	var nodeC = make(chan *Node)
	// start generator
	go func() {
		defer close(nodeC)
		for _, idx := range order {
			nodeC <- &g.Nodes[idx]
		}
	}()

	var wg sync.WaitGroup
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func() {
			worker(context.Background(), nodeC)
			wg.Done()
		}()
	}
	wg.Wait()
}

type Node struct {
	ID        string
	Delay     int
	ParentIDs []string

	wg       sync.WaitGroup // waitgroup is prerequisites wait group
	children []*Node        // addresses of all child nodes
}

func worker(ctx context.Context, nodeC <-chan *Node) {
	for {
		select {
		case <-ctx.Done():
			return
		case node, more := <-nodeC:
			if !more {
				return
			}
			node.wg.Wait()
			fmt.Println("start:", node.ID)
			<-time.After(time.Duration(node.Delay) * time.Second)
			fmt.Println("finish:", node.ID)
			for i := range node.children {
				node.children[i].wg.Done() // mark this prerequisite as done
			}
		}
	}
}

type Graph struct {
	Nodes []Node
	g     *graph.Mutable
}

func NewGraph(nodes []Node) Graph {
	var idx = make(map[string]int, len(nodes))
	for i := range nodes {
		idx[nodes[i].ID] = i
	}

	g := graph.New(len(nodes))
	for i := range nodes {
		fmt.Printf("node %d has %d parents\n", i, len(nodes[i].ParentIDs))
		nodes[i].wg.Add(len(nodes[i].ParentIDs))
		for _, parID := range nodes[i].ParentIDs {
			idxPar := idx[parID]
			g.Add(idxPar, i)
			nodes[idxPar].children = append(nodes[idxPar].children, &nodes[i])
		}
	}
	return Graph{Nodes: nodes, g: g}
}
