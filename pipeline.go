package dagproc

import (
	"errors"
	"fmt"

	"github.com/yourbasic/graph"
)

type pipeline struct {
	v []action
	g *graph.Mutable
}

// newPipeline creates a new pipeline.  It creates a  direct acyclic graph,
// with each "action" being a vertex, and dependencies represented by graph
// edges.
func newPipeline(nodes []Node) (pipeline, error) {
	// populate nodes index and actions.
	var (
		idx     = make(map[string]int, len(nodes))
		actions = make([]action, len(nodes))
	)
	for i := range nodes {
		// ensure that the id is unique.
		if _, seen := idx[nodes[i].ID()]; seen {
			return pipeline{}, fmt.Errorf("non-unique Node ID=%s", nodes[i].ID())
		}
		idx[nodes[i].ID()] = i
		actions[i] = action{
			n: nodes[i],
		}
	}

	// resolve dependencies
	g := graph.New(len(nodes))
	for i := range nodes {
		for _, parID := range nodes[i].ParentIDs() {
			// ensure that the parent ID exists.
			idxPar, exists := idx[parID]
			if !exists {
				return pipeline{}, fmt.Errorf("node ID=%s references non-existing node ID=%s", nodes[i].ID(), parID)
			}
			actions[idxPar].AddChild(&actions[i])
			g.Add(idxPar, i)
		}
		actions[i].IncParentCount(len(nodes[i].ParentIDs()))
	}
	// ensure there are no circular dependencies.
	if !graph.Acyclic(g) {
		return pipeline{}, errors.New("graph is not acyclic")
	}
	return pipeline{v: actions, g: g}, nil
}
