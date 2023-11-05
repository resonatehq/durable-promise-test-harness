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
		return c.Search(ctx, op)
	case store.Get:
		return c.Get(ctx, op)
	case store.Create:
		return c.Create(ctx, op)
	case store.Cancel:
		return c.Cancel(ctx, op)
	case store.Resolve:
		return c.Resolve(ctx, op)
	case store.Reject:
		return c.Reject(ctx, op)
	default:
		panic(fmt.Sprintf("unknown operation: %d", op.API))
	}
}

func (c *Client) Search(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.SearchPromisesParams)
		if !ok {
			panic(ok)
		}
		return c.client.SearchPromises(ctx, input)
	}

	return invoke[[]openapi.Promise](ctx, op, call, []int{200})
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
		if !ok || input == nil || input.Id == nil {
			panic(ok)
		}
		return c.client.CreatePromise(ctx, *input.Id, *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Cancel(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.CancelPromiseRequest)
		if !ok {
			panic(ok)
		}
		// qq: is value arbitrary, why no id??
		return c.client.CancelPromise(ctx, "id", *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Resolve(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.ResolvePromiseRequest)
		if !ok {
			panic(ok)
		}
		// qq: is value arbitrary, why no id??
		return c.client.ResolvePromise(ctx, "id", *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Reject(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.RejectPromiseRequest)
		if !ok {
			panic(ok)
		}
		// qq: is value arbitraty, why no id??
		return c.client.RejectPromise(ctx, "id", *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
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
