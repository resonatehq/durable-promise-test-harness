package simulator

import (
	"log"
	"math/rand"
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
	// TODO: clean up resonate.db so doesn't affect stuff
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
			WithGenerator(NewGenerator(&GeneratorConfig{
				r:    rand.New(rand.NewSource(0)),
				Ids:  100,
				Data: 100,
			})),
			WithChecker(checker.NewChecker()),
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
	st := store.NewStore()
	ops := t.Generator.Generate(10)

	for _, op := range ops {
		st.Add(t.Client.Invoke(op))
	}

	// write to file - create format:
	// 1) event history
	// 2) performance charts
	log.Println(t.Checker.Timeline(st.History()))

	return t.Checker.Check(st.History())
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
