package simulator

type Mode int

type SimulationConfig struct {
	Addr        string
	NumClients  int
	NumRequests int
}
