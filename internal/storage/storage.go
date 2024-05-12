package storage

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	// mux *sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}
