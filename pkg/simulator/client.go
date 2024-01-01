package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type Client struct {
	ID     int
	client openapi.ClientInterface
}

func NewClient(id int, conn string) (*Client, error) {
	c, err := openapi.NewClientWithResponses(conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		ID:     id,
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

	return invoke[openapi.SearchPromisesResponseObj](ctx, op, call, []int{200})
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
		input, ok := op.Input.(*openapi.CreatePromiseJSONRequestBody)
		if !ok || input == nil {
			panic(ok)
		}
		return c.client.CreatePromise(ctx, nil, *input)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Cancel(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.CompletePromiseRequestWrapper)
		if !ok {
			panic(ok)
		}
		body, ok := input.Request.(*openapi.PatchPromisesIdJSONRequestBody)
		if !ok || body == nil {
			panic(ok)
		}
		return c.client.PatchPromisesId(ctx, *input.Id, nil, *body)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Resolve(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.CompletePromiseRequestWrapper)
		if !ok {
			panic(ok)
		}
		body, ok := input.Request.(*openapi.PatchPromisesIdJSONRequestBody)
		if !ok || body == nil {
			panic(ok)
		}
		return c.client.PatchPromisesId(ctx, *input.Id, nil, *body)
	}

	return invoke[openapi.Promise](ctx, op, call, []int{200, 201}) // 200 for idempotency
}

func (c *Client) Reject(ctx context.Context, op store.Operation) store.Operation {
	call := func() (*http.Response, error) {
		input, ok := op.Input.(*openapi.CompletePromiseRequestWrapper)
		if !ok {
			panic(ok)
		}
		body, ok := input.Request.(*openapi.PatchPromisesIdJSONRequestBody)
		if !ok || body == nil {
			panic(ok)
		}
		return c.client.PatchPromisesId(ctx, *input.Id, nil, *body)
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
			break
		}
	}

	return op
}
