package store

import (
	"fmt"
	"time"
)

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

func (o Operation) String() string {
	return fmt.Sprintf(
		"Operation(id=%d, clientId=%d, api=%d, input=%v, output=%v)",
		o.ID,
		o.ClientID,
		o.API,
		o.Input,
		o.Output,
	)
}
