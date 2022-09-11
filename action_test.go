package dagproc

import (
	"testing"
	"time"

	"github.com/rusq/dagproc/internal/fixtures"
)

func Test_action_Wait(t *testing.T) {
	tests := []struct {
		name        string
		numDeps     int
		wantTimeout bool
	}{
		{
			"wait waits for what should be waited for",
			1,
			true,
		},
		{
			"wait terminates when there's nothing to wait",
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a = action{
				n: fixtures.FakeNode{WantID: "fake"},
			}
			a.wg.Add(tt.numDeps)
			if timedOut := waitTimesOut(&a, 10*time.Millisecond); timedOut != tt.wantTimeout {
				t.Errorf("wg.Wait(): want=%v, got=%v", tt.wantTimeout, timedOut)
			}
		})
	}
}

func Test_action_IncParentCount(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"make sure the n is propagated to wg",
			args{3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a = action{
				n: fixtures.FakeNode{WantID: "fake"},
			}
			a.IncParentCount(tt.args.n)
			for i := 0; i < tt.args.n; i++ {
				a.wg.Done()
			}
			if waitTimesOut(&a.wg, 10*time.Millisecond) {
				t.Fatalf("wg still expects a Done() call")
			}
		})
	}
}

func Test_action_DecParentCount(t *testing.T) {
	var tests = []struct {
		name        string
		parentCount int
		wantTimeout bool
	}{
		{
			"no references left",
			1,
			false,
		},
		{
			"1 reference remaining",
			2,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a = action{
				n: fixtures.FakeNode{WantID: "fake"},
			}
			a.wg.Add(tt.parentCount)
			// decrease once
			a.DecParentCount("x")

			if timedOut := waitTimesOut(&a.wg, 10*time.Millisecond); timedOut != tt.wantTimeout {
				t.Errorf("wg.Wait(): want=%v, got=%v", tt.wantTimeout, timedOut)
			}
		})
	}
}

type waiter interface {
	Wait()
}

// waitTimesOut returns true, if timeout happened before wg.Wait() returned,
// and false if wg.Wait() completed.
func waitTimesOut(wg waiter, timeout time.Duration) bool {
	var ok = make(chan struct{}, 1)
	go func() {
		wg.Wait()
		close(ok)
	}()
	select {
	case <-time.After(timeout):
		return true
	case <-ok:
		return false
	}
}

func Test_action_AddChild(t *testing.T) {
	var fakeChild = action{n: fixtures.FakeNode{WantID: "child", WantParentIDs: nil, WantDoErr: nil}}
	type fields struct {
		n        Node
		children []*action
	}
	type args struct {
		a *action
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   action
	}{
		{
			"adding a child action",
			fields{n: fixtures.FakeNode{WantID: "parent"}},
			args{&fakeChild},
			action{children: []*action{&fakeChild}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &action{
				n:        tt.fields.n,
				children: tt.fields.children,
			}
			v.AddChild(tt.args.a)
		})
	}
}
