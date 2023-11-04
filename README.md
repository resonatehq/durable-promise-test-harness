# Durable Promise Test Harness 

Harness is a verification system that checks implementations for conformance to the [Durable Promise Specification](https://github.com/resonatehq/durable-promise). 

## Key Features
- In-depth validation and correctness checks for single client 
- Load test with multiple client processes
- Detailed reporting and visualizations

## Verifying a Durable Promise Implementation  

This project contains a full test suite using go:
```bash
DP_SERVER=http://0.0.0.0:8001/ go test -v ./test/...  
```

## How It Works 

In Harness, tests are called "simulations". A simulation is controlled by a program (the 'simulator'). The simulator launches clients and contains test logic. Results are aggregated for display in a web browser. 

1. A test runs as a Golang program. That program setups the implementation you're going to test. 

2. Once the *system* is running, the harness spins up a set of logically single-threaded processes, each with its own client for the distributed system. 

3. A *generator* generates new operations for each process to perform. 

4. Processes then apply those operations to the system using their clients. The start and end of each operation is recorded in a history. 

5. Implementations are torn down. 

6. Harness uses a *checker* to analyze the test's history for correctnes and to generate benchkmark reports & timeline graphs. 

NOTE: the test, history, analysis, and any supplementary results are written to the filesystem under store/<test-name>/<date> for later review. Symlinks to the latest results are maintained at each level for convenience. 

## Contributions

We welcome bug reports, feature requests, and pull requests!

Before submitting a PR, please make sure:

- New tests are included
- All tests are passing
- Code is properly formatted
- Documentation is updated

