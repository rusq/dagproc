package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/rusq/dagproc"
)

var _ dagproc.Node = &DelayNode{}

// DelayNode is a simple node that, when executed, delays for a number of
// `delay` seconds.
type DelayNode struct {
	id        string
	parentIDs []string
	delay     int
}

func (n DelayNode) ID() string {
	return n.id
}

func (n DelayNode) ParentIDs() []string {
	return n.parentIDs
}

func (n DelayNode) Do() error {
	time.Sleep(time.Duration(n.delay) * time.Second)
	return nil
}

var numWorkers = flag.Uint("w", 4, "`number` of workers (must be > 0)")

func main() {
	flag.Parse()
	if *numWorkers == 0 {
		*numWorkers = 4
	}

	var nodes = []dagproc.Node{
		DelayNode{id: "0", delay: 1, parentIDs: []string{}},
		DelayNode{id: "1", delay: 1, parentIDs: []string{"0"}},
		DelayNode{id: "2", delay: 1, parentIDs: []string{"1"}},
		DelayNode{id: "3", delay: 4, parentIDs: []string{"1"}},
		DelayNode{id: "4", delay: 1, parentIDs: []string{"2"}},
		DelayNode{id: "5", delay: 1, parentIDs: []string{"4"}},
		DelayNode{id: "6", delay: 2, parentIDs: []string{"2", "3"}},
		DelayNode{id: "7", delay: 1, parentIDs: []string{"1", "3", "8"}},
		DelayNode{id: "8", delay: 1, parentIDs: []string{"5", "6"}},
	}
	if err := dagproc.Process(context.Background(), nodes, int(*numWorkers)); err != nil {
		log.Fatal(err)
	}
}
