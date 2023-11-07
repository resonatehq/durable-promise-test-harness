package simulator

import (
	"encoding/base64"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

type GeneratorConfig struct {
	r *rand.Rand

	numRequests int

	Ids  int
	Data int
}

type Generator struct {
	r           *rand.Rand
	numRequests int
	idSet       []string
	dataSet     [][]byte
}

func NewGenerator(config *GeneratorConfig) *Generator {
	idSet := make([]string, config.Ids)
	for i := 0; i < config.Ids; i++ {
		idSet[i] = strconv.Itoa(i)
	}

	dataSet := [][]byte{}
	for i := 0; i < config.Data; i++ {
		dataSet = append(dataSet, []byte(strconv.Itoa(i)), nil) // half of all values are nil
	}

	return &Generator{
		r:           config.r,
		numRequests: config.numRequests,
		idSet:       idSet,
		dataSet:     dataSet,
	}
}

func (g *Generator) Generate() []store.Operation {
	ops := []store.Operation{}

	clientId := int(uuid.New().ID())

	generators := []OpGenerator{
		g.GenerateSearchPromise,
		g.GenerateReadPromise,
		g.GenerateCreatePromise,
		g.GenerateCancelPromise,
		g.GenerateResolvePromise,
		g.GenerateRejectPromise,
	}

	for i := 0; i < g.numRequests; i++ {
		bound := len(generators)
		ops = append(ops, generators[g.r.Intn(bound)](g.r, clientId))
	}

	return ops
}

type OpGenerator func(*rand.Rand, int) store.Operation

func (g *Generator) GenerateSearchPromise(r *rand.Rand, clientID int) store.Operation {
	// state must be one of: pending, resolved, rejected
	var state string
	switch r.Intn(3) {
	case 0:
		state = string(openapi.PENDING)
	case 1:
		state = string(openapi.RESOLVED)
	case 2:
		// rejected gets the timeout and canceled promises too
		state = string(openapi.REJECTED)
	}

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Search,
		Input: &openapi.SearchPromisesParams{
			Q:     "*",
			State: utils.ToPointer(state),
			// Limit: utils.ToPointer(),
			// Cursor: utils.ToPointer(),
		},
	}
}

func (g *Generator) GenerateReadPromise(r *rand.Rand, clientID int) store.Operation {
	promiseId := g.idSet[r.Intn(len(g.idSet))]

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Get,
		Input:    promiseId,
	}
}

func (g *Generator) GenerateCreatePromise(r *rand.Rand, clientID int) store.Operation {
	promiseId := g.idSet[r.Intn(len(g.idSet))]
	data := g.dataSet[r.Intn(len(g.dataSet))]
	// timeout := r.Int63n(max-min) + min

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Create,
		Input: &openapi.CreatePromiseRequest{
			Id: utils.ToPointer(promiseId),
			Param: &openapi.Value{
				Data: utils.ToPointer(base64.StdEncoding.EncodeToString(data)),
			},
			Timeout: utils.ToPointer(2524608000000),
		},
	}
}

func (g *Generator) GenerateCancelPromise(r *rand.Rand, clientID int) store.Operation {
	promiseId := g.idSet[r.Intn(len(g.idSet))]
	data := g.dataSet[r.Intn(len(g.dataSet))]

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Cancel,
		Input: &openapi.CompletePromiseRequestWrapper{
			Id: utils.ToPointer(promiseId),
			Request: &openapi.CancelPromiseRequest{
				Value: &openapi.Value{
					Data: utils.ToPointer(base64.StdEncoding.EncodeToString(data)),
				},
			},
		},
	}
}

func (g *Generator) GenerateResolvePromise(r *rand.Rand, clientID int) store.Operation {
	promiseId := g.idSet[r.Intn(len(g.idSet))]
	data := g.dataSet[r.Intn(len(g.dataSet))]

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Resolve,
		Input: &openapi.CompletePromiseRequestWrapper{
			Id: utils.ToPointer(promiseId),
			Request: &openapi.ResolvePromiseRequest{
				Value: &openapi.Value{
					Data: utils.ToPointer(base64.StdEncoding.EncodeToString(data)),
				},
			},
		},
	}
}

func (g *Generator) GenerateRejectPromise(r *rand.Rand, clientID int) store.Operation {
	promiseId := g.idSet[r.Intn(len(g.idSet))]
	data := g.dataSet[r.Intn(len(g.dataSet))]

	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Reject,
		Input: &openapi.CompletePromiseRequestWrapper{
			Id: utils.ToPointer(promiseId),
			Request: &openapi.RejectPromiseRequest{
				Value: &openapi.Value{
					Data: utils.ToPointer(base64.StdEncoding.EncodeToString(data)),
				},
			},
		},
	}
}
