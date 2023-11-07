package simulator

type Mode int

const (
	Linearizability Mode = iota
	Load
)

type SimulationConfig struct {
	Addr        string
	NumClients  int
	NumRequests int
	Mode        Mode
}
