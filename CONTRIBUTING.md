# Contributing to durable-promise-test-harness

Thank you for your interest in contributing to the Durable Promise Test Harness!
This guide will get you from zero to an open pull request.

---

## What is this project?

The **Durable Promise Test Harness** is a conformance-testing tool for implementations
of the [Durable Promise Specification](https://github.com/resonatehq/durable-promise).
It stress-tests a running Durable Promise server by generating pseudo-random operation
sequences and verifying that the observed history is linearizable using
[Porcupine](https://github.com/anishathalye/porcupine).

### Key components

| Component | Location | Role |
|---|---|---|
| Simulator | `pkg/simulator/` | Generates randomized client operations |
| LocalStore | `pkg/store/` | Records operation start/end times |
| Checker | `pkg/checker/` | Verifies linearizability of the history |
| CLI | `cmd/` | Cobra-based entry point (`harness verify`) |

---

## Prerequisites

- **Go 1.21+**
- A running Durable Promise server (see [durable-promise](https://github.com/resonatehq/durable-promise) for reference implementations, or a local [Resonate server](https://github.com/resonatehq/resonate))

---

## Development setup

```bash
# Fork the repo on GitHub, then clone your fork
git clone https://github.com/your-username/durable-promise-test-harness
cd durable-promise-test-harness

go mod download
go build -o harness
go test ./...
```

---

## Running the harness

Start your Durable Promise server on port 8001, then run:

```bash
./harness verify -a http://0.0.0.0:8001/ -r 1000 -c 3
```

| Flag | Description |
|------|-------------|
| `-a` | Address of the Durable Promise server (required) |
| `-r` | Total number of requests to generate (required) |
| `-c` | Number of concurrent clients (required) |

Results are written to `test/results/<date>/` for later review.

---

## Project structure

```
.
├── cmd/verify/         # CLI command implementation
├── pkg/
│   ├── checker/        # Linearizability checker (Porcupine wrapper)
│   ├── openapi/        # HTTP client generated from OpenAPI spec
│   ├── simulator/      # Operation generator and simulation loop
│   ├── store/          # In-memory operation history store
│   └── utils/
├── main.go
├── go.mod
└── Makefile
```

---

## Submitting a pull request

1. Browse [open issues](https://github.com/resonatehq/durable-promise-test-harness/issues).
   Issues labelled **good first issue** are a great place to start.
2. Comment on the issue to let maintainers know you are working on it.
3. Create a branch from `main`:
   ```bash
   git checkout -b your-username/short-description
   ```
4. Make your change. Keep commits focused and atomic.
5. Add or update tests. Run `go test ./...` before pushing.
6. Format your code: `gofmt -w . && go vet ./...`
7. Push and open a PR against `main`.

### PR checklist

- [ ] Tests added or updated
- [ ] All tests pass (`go test ./...`)
- [ ] Code formatted (`gofmt -w .`)
- [ ] Documentation updated if behaviour changed

---

## Code style

- Follow standard Go conventions (see [Effective Go](https://go.dev/doc/effective_go)).
- Keep functions small and names descriptive.
- Add comments for exported symbols.

---

## Getting help

- [Open issues](https://github.com/resonatehq/durable-promise-test-harness/issues)
- [Durable Promise Specification](https://github.com/resonatehq/durable-promise)
- Open a new issue if you hit a problem not already tracked

We look forward to your contribution!
