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

	Ids  int
	Data int
}

// Generator is responsible for telling the test what operations to perform
// on the implementation during the test. It outputs a sequence of operations.
type Generator struct {
	r       *rand.Rand
	idSet   []string
	dataSet [][]byte
}

// NewGenerator return a set of generate dataset used to create operations.
func NewGenerator(config *GeneratorConfig) *Generator {
	idSet := make([]string, config.Ids)
	for i := 0; i < config.Ids; i++ {
		idSet[i] = strconv.Itoa(i)
	}

	dataSet := [][]byte{}
	for i := 0; i < config.Data; i++ {
		dataSet = append(dataSet, []byte(strconv.Itoa(i)), nil) // half of all values are nil ???
	}

	return &Generator{
		r:       config.r,
		idSet:   idSet,
		dataSet: dataSet,
	}
}

func (g *Generator) Generate(n int) []store.Operation {
	ops := []store.Operation{}

	clientId := int(uuid.New().ID())

	generators := []OpGenerator{
		// g.GenerateSearchPromise,
		g.GenerateReadPromise,
		g.GenerateCreatePromise,
		g.GenerateCancelPromise,
		g.GenerateResolvePromise,
		g.GenerateRejectPromise,
	}

	// TODO: make generator more interesting and more likely to pick from the created promises...?
	for i := 0; i < n; i++ {
		bound := len(generators)
		ops = append(ops, generators[g.r.Intn(bound)](g.r, clientId))
	}

	return ops
}

type OpGenerator func(*rand.Rand, int) store.Operation

func (g *Generator) GenerateSearchPromise(r *rand.Rand, clientID int) store.Operation {
	return store.Operation{
		ID:       int(uuid.New().ID()),
		ClientID: clientID,
		API:      store.Search,
		Input:    &openapi.SearchPromisesParams{},
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
			// ISSUE FIX -- even in curl...
			// TODO: nil makes it set to 0, whichs makes it time out immediately
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
