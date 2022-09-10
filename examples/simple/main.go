package main

import (
	"log"
	"time"

	"github.com/rusq/dagproc"
)

var _ dagproc.Node = &Node{}

type Node struct {
	id        string
	delay     int
	parentIDs []string
}

func (n Node) ID() string {
	return n.id
}

func (n Node) ParentIDs() []string {
	return n.parentIDs
}

func (n Node) Do() error {
	time.Sleep(time.Duration(n.delay) * time.Second)
	return nil
}

func main() {
	var nodes = []dagproc.Node{
		Node{id: "0", delay: 1, parentIDs: []string{}},
		Node{id: "1", delay: 1, parentIDs: []string{"0"}},
		Node{id: "2", delay: 1, parentIDs: []string{"1"}},
		Node{id: "3", delay: 4, parentIDs: []string{"1"}},
		Node{id: "4", delay: 1, parentIDs: []string{"2"}},
		Node{id: "5", delay: 1, parentIDs: []string{"4"}},
		Node{id: "6", delay: 2, parentIDs: []string{"2", "3"}},
		Node{id: "7", delay: 1, parentIDs: []string{"1", "3", "8"}},
		Node{id: "8", delay: 1, parentIDs: []string{"5", "6"}},
	}
	// var nodes = []dagproc.Node{
	// 	Node{id: "0", delay: 1},
	// 	Node{id: "1", delay: 1},
	// 	Node{id: "2", delay: 1},
	// 	Node{id: "3", delay: 1, parentIDs: []string{"0", "1", "2"}},
	// }
	if err := dagproc.Process(nodes, 4); err != nil {
		log.Fatal(err)
	}
}
