package dagproc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/yourbasic/graph"
)

type Node interface {
	ID() string
	ParentIDs() []string
	Do() error
}

type Graph struct {
	v []vertex
	g *graph.Mutable
}

func NewGraph(nodes []Node) (Graph, error) {
	var (
		idx      = make(map[string]int, len(nodes))
		vertices = make([]vertex, len(nodes))
	)
	for i := range nodes {
		idx[nodes[i].ID()] = i
		vertices[i] = vertex{
			n: nodes[i],
		}
	}
	// add vertex edges
	g := graph.New(len(nodes))
	for i := range nodes {
		for _, parID := range nodes[i].ParentIDs() {
			idxPar := idx[parID]
			g.Add(idxPar, i)
			vertices[idxPar].children = append(vertices[idxPar].children, &vertices[i])
		}
		vertices[i].setParentCount(len(nodes[i].ParentIDs()))
	}
	if !graph.Acyclic(g) {
		return Graph{}, errors.New("graph is not acyclic")
	}
	return Graph{v: vertices, g: g}, nil
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
			worker(context.Background(), i, vertexC)
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

func (v *vertex) setParentCount(n int) {
	log.Printf("node %s: set parent count: %d", v.n.ID(), n)
	v.wg.Add(n)
}

func (v *vertex) reportOK(id string) {
	log.Printf("node %s: parent %s reported OK", v.n.ID(), id)
	v.wg.Done()
}

func worker(ctx context.Context, id int, vertexC <-chan *vertex) {
	lg := log.New(os.Stdout, "worker "+strconv.Itoa(id)+": ", log.LstdFlags)
	for {
		select {
		case <-ctx.Done():
			return
		case vertex, more := <-vertexC:
			if !more {
				return
			}
			lg.Println("waiting on dependencies", vertex.n.ID())
			vertex.wg.Wait()
			lg.Println("start:", vertex.n.ID(), "num children:", len(vertex.children))
			if err := vertex.n.Do(); err != nil {
				log.Printf("node: %s, error: %s", vertex.n.ID(), err)
			}
			lg.Println("finish:", vertex.n.ID())
			for i := range vertex.children {
				lg.Printf("notify %s that vertex %s is done", vertex.children[i].n.ID(), vertex.n.ID())
				vertex.children[i].reportOK(vertex.n.ID()) // mark this prerequisite as done
			}
		}
	}
}
