package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type Client struct {
	ID     int
	client openapi.ClientInterface
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

// Invoke receives the start of an operation and returns the end of it
func (c *Client) Invoke(ctx context.Context, op store.Operation) store.Operation {
	switch op.API {
	case store.Search:
		return c.Search(op)
	case store.Get:
		return c.Get(ctx, op)
	case store.Create:
		return c.Create(ctx, op)
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

func (c *Client) Search(op store.Operation) store.Operation {
	return store.Operation{
		Status: store.Ok,
		API:    store.Search,
		Output: "TODO",
	}
}

func (c *Client) Get(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(string)
		if !ok {
			panic(ok)
		}
		return c.client.GetPromise(ctx, input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200})
}

func (c *Client) Create(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.CreatePromiseRequest)
		if !ok || input == nil {
			panic(ok)
		}
		return c.client.CreatePromise(ctx, *input.Id, *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201})
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

func invoke[T any](ctx context.Context, op store.Operation, call func() (*http.Response, error), ok []int) store.Operation {
	op.CallEvent = time.Now()

	resp, err := call()
	if err != nil {
		op.ReturnEvent = time.Now()
		return op
	}

	op.ReturnEvent = time.Now()

	op.Code = resp.StatusCode

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return op
	}

	var out T
	err = json.Unmarshal(b, &out)
	if err != nil {
		return op
	}

	op.Output = &out

	op.Status = store.Fail
	for i := range ok {
		if ok[i] == op.Code {
			op.Status = store.Ok
		}
	}

	return op
}
