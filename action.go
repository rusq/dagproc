package dagproc

import (
	"sync"

	"github.com/rusq/dlog"
)

type action struct {
	n        Node
	wg       sync.WaitGroup // waitgroup is prerequisites wait group
	children []*action      // addresses of all child nodes
}

// Wait blocks until all parent tasks are complete.
func (v *action) Wait() {
	dlog.Debugf("action id=%s: waiting for parent tasks to complete")
	v.wg.Wait()
}

// IncParentCount increases the parent count.
func (v *action) IncParentCount(n int) {
	dlog.Debugf("action id=%s: set parent count: %d", v.n.ID(), n)
	v.wg.Add(n)
}

// DecParentCount reports that parent process with ID is completed.  id is used
// for debugging.
func (v *action) DecParentCount(id string) {
	dlog.Debugf("action id=%s: parent %s finished", v.n.ID(), id)
	v.wg.Done()
}

func (v *action) AddChild(a *action) {
	dlog.Debugf("action id=%s: adding child action id=%s", v.n.ID(), a.n.ID())
	v.children = append(v.children, a)
}
