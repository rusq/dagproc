package fixtures

type FakeNode struct {
	WantID        string
	WantParentIDs []string
	WantDoErr     error
}

func (f FakeNode) ID() string {
	return f.WantID
}

func (f FakeNode) ParentIDs() []string {
	return f.WantParentIDs
}

func (f FakeNode) Do() error {
	return f.WantDoErr
}
