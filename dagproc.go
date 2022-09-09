package dagproc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/yourbasic/graph"
)

type Node interface {
	ID() string
	ParentIDs() []string
	Do() error
}

func Process(n []Node, workers int) error {
	g, err := NewGraph(n)
	if err != nil {
		return err
	}
	order, ok := graph.TopSort(g.g)
	if !ok {
		log.Fatal("poop")
	}

	var vertexC = make(chan *vertex)
	// start generator
	go func() {
		defer close(vertexC)
		for _, idx := range order {
			vertexC <- &g.v[idx]
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			worker(context.Background(), vertexC)
			wg.Done()
			fmt.Printf("worker %d exited\n", i)
		}(i)
	}
	wg.Wait()
	return nil
}

type vertex struct {
	n        Node
	wg       sync.WaitGroup // waitgroup is prerequisites wait group
	children []*vertex      // addresses of all child nodes
}

func worker(ctx context.Context, vertexC <-chan *vertex) {
	for {
		select {
		case <-ctx.Done():
			return
		case vertex, more := <-vertexC:
			if !more {
				return
			}
			fmt.Println("waiting on dependencies", vertex.n.ID())
			vertex.wg.Wait()
			fmt.Println("start:", vertex.n.ID())
			if err := vertex.n.Do(); err != nil {
				log.Printf("node: %s, error: %s", vertex.n.ID(), err)
			}
			fmt.Println("finish:", vertex.n.ID())
			for i := range vertex.children {
				vertex.children[i].wg.Done() // mark this prerequisite as done
			}
		}
	}
}

type Graph struct {
	v []vertex
	g *graph.Mutable
}

func NewGraph(nodes []Node) (Graph, error) {
	var idx = make(map[string]int, len(nodes))
	for i := range nodes {
		idx[nodes[i].ID()] = i
	}

	var vertices = make([]vertex, len(nodes))

	g := graph.New(len(nodes))
	for i := range nodes {
		vertices[i] = vertex{
			n: nodes[i],
		}
		fmt.Printf("node %d has %d parents\n", i, len(nodes[i].ParentIDs()))
		vertices[i].wg.Add(len(nodes[i].ParentIDs()))
		for _, parID := range nodes[i].ParentIDs() {
			idxPar := idx[parID]
			g.Add(idxPar, i)
			vertices[idxPar].children = append(vertices[idxPar].children, &vertices[i])
		}
	}
	if !graph.Acyclic(g) {
		return Graph{}, errors.New("graph is not acyclic")
	}
	return Graph{v: vertices, g: g}, nil
}
