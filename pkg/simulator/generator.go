package simulator

import (
	"github.com/google/uuid"
	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

// Generator is responsible for telling the test what operations to perform
// on the implementation during the test. It outputs a sequence of operations.
type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() []store.Operation {
	// ops := make([]store.Operation, 0)

	clientId := int(uuid.New().ID())
	ops := []store.Operation{
		{
			ID:       int(uuid.New().ID()),
			ClientID: clientId,
			API:      store.Create,
			Input: &openapi.CreatePromiseRequest{
				Id:    utils.ToPointer("foo"),
				Param: &openapi.Value{
					// Data: utils.ToPointer(`"'"$(echo -n 'Durable Promise Resolved' | base64)"'"`), // TODO: understand more here
				},
				Timeout: utils.ToPointer(2524608000000),
			},
		},
		{
			ID:       int(uuid.New().ID()),
			ClientID: clientId,
			API:      store.Create,
			Input: &openapi.CreatePromiseRequest{
				Id:    utils.ToPointer("bar"), // id is userdefined
				Param: &openapi.Value{
					// Data: utils.ToPointer(`"'"$(echo -n 'Durable Promise Resolved' | base64)"'"`), // TODO: understand more here
				},
				Timeout: utils.ToPointer(2524608000000),
			},
		},
		{
			ID:       int(uuid.New().ID()),
			ClientID: clientId,
			API:      store.Get,
			Input:    "foo",
		},
		{
			ID:       int(uuid.New().ID()),
			ClientID: clientId,
			API:      store.Get,
			Input:    "bar",
		},
	}

	return ops
}

// TODO: look at dst for inspiration
