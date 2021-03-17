package quote

type Quote struct {
	ID     int64
	Symbol string
	Name   string
	Status Status
}

type Status string

const (
	StatusNew   Status = "new"
	StatusReady Status = "ready"
)
