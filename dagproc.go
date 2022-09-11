package dagproc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rusq/dlog"
	"github.com/yourbasic/graph"
)

// defNumWorkers is the default number of workers if caller requests 0 workers.
const defNumWorkers = 1

// A Node is a basic building block of a pipeline.
//
//go:generate mockgen -destination internal/mock_dagproc/mock_dagproc.go . Node
type Node interface {
	// ID should return an ID, i.e. "3".  ID should be unique across all
	// nodes.
	ID() string
	// ParentIDs should return a slice of IDs that this node depends on, i.e.
	// ["1", "2"].  There should be no cyclic references.
	ParentIDs() []string
	// Do should execute the node action and return an error.  If it returns
	// [ErrIgnore], the error will be ignored, with the message in [ErrIgnore]
	// (you can use the [NewErrIgnore] to initialise it). Any other error type
	// will lead to worker and process termination.
	Do() error
}

// ErrProcess is a Process error.
type ErrProcess struct {
	// NodeID is a node ID that reported an error
	NodeID string
	// Err is an actual error, reported by a node.
	Err error
	// Worker is the id of the worker that reported this error.
	Worker int
	// Time when the error occurred.
	Time time.Time
}

func (ep *ErrProcess) Error() string {
	return fmt.Sprintf("an error occurred at %s on worker: %d node: %s error: %s", ep.Time, ep.Worker, ep.NodeID, ep.Err)
}

// ErrIgnore should be returned by [Node] if the error should be ignored and the
// process should continue running.  It can be instantiated with
// [NewErrIgnore].
type ErrIgnore string

// NewErrIgnore instantiates a new [ErrIgnore] error.
func NewErrIgnore(msg string) error {
	return ErrIgnore(msg)
}

func (e ErrIgnore) Error() string {
	return "process error (ignored): " + string(e)
}

// ErrNothingToDo is returned by [Process] if there are no Nodes.
var ErrNothingToDo = errors.New("nothing to do")

// Process processes a slice of nodes with requested number of workers.
// workers defaults to defNumWorkers if 0 workers is requested.  If n does not
// contain any nodes, Process returns [ErrNothingToDo].
func Process(ctx context.Context, n []Node, numWorkers int) error {
	if len(n) == 0 {
		return ErrNothingToDo
	}
	if numWorkers <= 0 {
		numWorkers = defNumWorkers
	}
	g, err := newPipeline(n)
	if err != nil {
		return fmt.Errorf("error initialising DAG: %w", err)
	}
	order, ok := graph.TopSort(g.g)
	if !ok {
		return errors.New("topological sort failed on the graph")
	}

	var actionC = make(chan *action)
	// start a generator
	go func() {
		defer close(actionC)
		for _, idx := range order {
			actionC <- &g.v[idx]
		}
	}()

	// start workers
	var wg sync.WaitGroup
	var errC = make(chan error, 1)
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(i int) {
			worker(ctx, i, actionC, errC)
			wg.Done()
			fmt.Printf("worker %d exited\n", i)
		}(i)
	}
	go func() {
		wg.Wait()
		close(errC)
	}()
	// process worker errors, if any.
	for err := range errC {
		if err != nil {
			return err
		}
	}
	return nil
}

// worker is the function that executes Node actions.  It receives actions on
// actionC channel.  If an error occurs and it's not an ErrIgnore, it sends it
// on errC and terminates.  If it is an ErrIgnore error, it logs it with the
// context logger and continues execution.
func worker(ctx context.Context, id int, actionC <-chan *action, errC chan<- error) {
	lg := dlog.FromContext(ctx)
	lg.SetPrefix("worker " + strconv.Itoa(id) + " ")
	for {
		select {
		case <-ctx.Done():
			return
		case action, more := <-actionC:
			if !more {
				return
			}
			lg.Debugln("waiting on dependencies of", action.n.ID())
			action.Wait()
			lg.Debugf("action id=%s: started, num children: %d", action.n.ID(), len(action.children))
			if err := action.n.Do(); err != nil {
				ep := &ErrProcess{NodeID: action.n.ID(), Worker: id, Err: err, Time: time.Now()}
				// we got to return if it's not an "ignore" error.
				if !isIgnoreError(err) {
					lg.Debugf("worker %d exiting with error: %s", id, err)
					errC <- ep
					return
				}
				lg.Print(err)
			}
			lg.Debugln("finish:", action.n.ID())
			for i := range action.children {
				lg.Debugf("notify %s that action id=%s is done", action.children[i].n.ID(), action.n.ID())
				action.children[i].DecParentCount(action.n.ID()) // mark this prerequisite as done
			}
		}
	}
}

// isIgnoreError returns true if the err is ErrIgnore.
func isIgnoreError(err error) bool {
	var ei ErrIgnore
	return errors.As(err, &ei)
}
