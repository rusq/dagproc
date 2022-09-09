package main

import (
	"fmt"
	"log"

	"github.com/yourbasic/graph"
)

func main() {
	graphtest()
}

func graphtest() {
	g := graph.New(9)
	g.AddCost(0, 1, 1)
	g.AddCost(1, 2, 1)
	g.AddCost(1, 3, 4)
	g.AddCost(1, 7, 1)
	g.AddCost(3, 6, 1)
	g.AddCost(3, 7, 1)
	g.AddCost(2, 4, 1)
	g.AddCost(2, 6, 1)
	g.AddCost(6, 8, 2)
	g.AddCost(4, 5, 1)
	g.AddCost(5, 8, 1)
	g.AddCost(8, 7, 1)

	if !graph.Acyclic(g) {
		log.Fatal("cyclic graph")
	}
	fmt.Println(graph.String(g))

	fmt.Println("** BFS **")
	var i = 0
	graph.BFS(g, 0, func(v, w int, c int64) {
		fmt.Printf("%d: v=%d, w=%d, c=%d\n", i, v, w, c)
		i++
	})

	fmt.Println("** Components **: ", graph.Components(g))

	fmt.Println("** MST **")
	enumerate(graph.MST(g))

	fmt.Println("** Sorted **")
	sg := graph.Sort(g)
	fmt.Println(sg.String())

	fmt.Println("** TopSort **:")
	fmt.Println(graph.TopSort(g))

}

func enumerate(a []int) {
	for i, v := range a {
		fmt.Printf("%d: %d\n", i, v)
	}
}
