package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type Client struct {
	ID     int
	client openapi.ClientWithResponsesInterface
}

func NewClient(conn string) (*Client, error) {
	c, err := openapi.NewClientWithResponses(conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		ID:     rand.Intn(500),
		client: c,
	}, nil
}

// Invoke recieves the start of an operation and returns the end of it
func (c *Client) Invoke(op store.Operation) store.Operation {
	switch op.API {
	case store.Search:
		return c.Search(op)
	case store.Get:
		return c.Get(op)
	case store.Create:
		return c.Create(op)
	case store.Cancel:
		return c.Cancel(op)
	case store.Resolve:
		return c.Resolve(op)
	case store.Reject:
		return c.Reject(op)
	default:
		panic(fmt.Sprintf("unknown operation: %d", op.API))
	}
}

//
// implement client for durable promise specification - write/read operations
// nil for read operations
//

func (c *Client) Search(op store.Operation) store.Operation {
	return store.Operation{
		Status: store.Ok,
		API:    store.Search,
		Output: "TODO",
	}
}

// TODO: use openapi for less boiler plate
func (c *Client) Get(op store.Operation) store.Operation {
	end := op
	ctx := context.Background()

	end.CallEvent = time.Now()
	input, ok := op.Input.(string)
	if !ok {
		panic(ok)
	}

	resp, err := c.client.GetPromiseWithResponse(ctx, input)
	if err != nil {
		panic(err)
	}
	end.ReturnEvent = time.Now()

	var out openapi.Promise
	err = json.Unmarshal(resp.Body, &out)
	if err != nil {
		panic(err)
	}

	end.Output = &out
	end.Status = store.Ok // validate before ?

	return end
}

func (c *Client) Create(op store.Operation) store.Operation {
	end := op
	ctx := context.Background()

	end.CallEvent = time.Now()

	input, ok := op.Input.(*openapi.CreatePromiseRequest)
	if !ok {
		panic(ok)
	}

	bs, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	reader := bytes.NewReader(bs)

	resp, err := c.client.CreatePromiseWithBodyWithResponse(ctx, *input.Id, "application/json", reader)
	if err != nil {
		panic(err)
	}

	end.ReturnEvent = time.Now()

	// TODO: CHECK STATUS BEFORE ANYTHING

	var out openapi.Promise
	err = json.Unmarshal(resp.Body, &out)
	if err != nil {
		panic(err)
	}

	end.Output = &out
	end.Status = store.Ok // validate before here ?

	return end
}

func (c *Client) Cancel(op store.Operation) store.Operation {
	return store.Operation{
		Status: store.Invoke,
		API:    store.Cancel,
		Output: "TODO",
	}
}

func (c *Client) Resolve(op store.Operation) store.Operation {
	return store.Operation{
		Status: store.Invoke,
		API:    store.Resolve,
		Output: "TODO",
	}
}

func (c *Client) Reject(op store.Operation) store.Operation {
	return store.Operation{
		Status: store.Invoke,
		API:    store.Reject,
		Output: "TODO",
	}
}
