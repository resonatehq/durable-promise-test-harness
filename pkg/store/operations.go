package store

import "time"

type Status int

const (
	Invoke Status = iota
	Ok
	Fail
)

type API int

const (
	Search API = iota
	Get
	Create
	Cancel
	Resolve
	Reject
)

// Operation is an element of a history.
type Operation struct {
	ID          int
	ClientID    int
	API         API
	Input       interface{}
	Output      interface{}
	CallEvent   time.Time
	ReturnEvent time.Time
	Status      Status
}
