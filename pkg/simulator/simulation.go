package simulator

import (
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/checker"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Simulation is a client
type Simulation struct {
	suite.Suite

	config *SimulationConfig
}

func New(config *SimulationConfig) *Simulation {
	// TODO: parse + validate
	return &Simulation{
		config: config,
	}
}

func (s *Simulation) SetupSuite() {
	s.T().Logf("testing server readiness at %s\n", s.config.Addr)

	var ready bool
	for i := 0; i < 10; i++ {
		if IsReady(s.config.Addr) {
			ready = true
			break
		}

		time.Sleep(3 * time.Second)
	}

	if !ready {
		s.T().Fatal("server did not become ready in time.")
	}
}

func (s *Simulation) TearDownSuite() {
	// TODO: delete everything created, leave server in original state
}

func (s *Simulation) TestSingleClientCorrectness() {
	defer func() {
		if r := recover(); r != nil {
			s.TearDownSuite()
			panic(r)
		}
	}()

	s.T().Run("single client correctness", func(t *testing.T) {
		client, err := NewClient(s.config.Addr)
		if err != nil {
			panic(err)
		}

		test := NewTest(
			WithClient(client),
			WithGenerator(NewGenerator()),
			WithChecker(checker.New()),
		)

		assert.Nil(t, test.Run()) // output: server logs, checker data
	})
}

type Test struct {
	Name      string
	Client    *Client
	Generator *Generator
	Checker   *checker.Checker
}

type TestOption func(*Test)

func NewTest(opts ...TestOption) *Test {
	test := &Test{}
	for _, opt := range opts {
		opt(test)
	}
	return test
}

func WithClient(c *Client) TestOption {
	return func(t *Test) {
		t.Client = c
	}
}

func WithGenerator(g *Generator) TestOption {
	return func(t *Test) {
		t.Generator = g
	}
}

func WithChecker(c *checker.Checker) TestOption {
	return func(t *Test) {
		t.Checker = c
	}
}

func (t *Test) Run() error {
	log.Printf("running tests...\n")

	s := store.NewStore()

	ops := t.Generator.Generate()

	for _, op := range ops {
		s.Add(t.Client.Invoke(op))
	}

	return t.Checker.Check(s.History())
}

func IsReady(Addr string) bool {
	// TODO: Addr define the format -- options -- remote local etc, tcp though
	serverAddr := strings.TrimSuffix(strings.TrimPrefix(Addr, "http://"), "/")
	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
