package mocks

import (
	"errors"
	"sync"

	"github.com/collapsinghierarchy/encproc/models"
)

// Use a mutex to allow safe concurrent access in tests
var mu sync.Mutex

// mockEncProc holds the latest inserted/updated values
var mockEncProc = models.EncProc{
	ID:     "",
	Pk:     nil,
	Params: "",
}

type EncProcModel struct{}

// InsertAggregationParams mock: stores values in mockEncProc
func (m *EncProcModel) InsertAggregationParams(id string, pk []byte, params string) error {
	mu.Lock()
	defer mu.Unlock()
	mockEncProc.ID = id
	mockEncProc.Pk = pk
	mockEncProc.Params = params
	return nil
}

// UpdateAggregationParams mock: updates values in mockEncProc if ID matches
func (m *EncProcModel) UpdateAggregationParams(id string, pk []byte, params string) error {
	mu.Lock()
	defer mu.Unlock()
	if mockEncProc.ID != id {
		return errors.New("not found")
	}
	mockEncProc.Pk = pk
	mockEncProc.Params = params
	return nil
}

// DeleteAggregationParams mock: clears mockEncProc if ID matches
func (m *EncProcModel) DeleteAggregationParams(id string) error {
	mu.Lock()
	defer mu.Unlock()
	if mockEncProc.ID == id {
		mockEncProc = models.EncProc{}
	}
	return nil
}

// GetAggregationParamsByID mock: returns stored values if ID matches
func (m *EncProcModel) GetAggregationParamsByID(id string) (string, []byte, string, error) {
	mu.Lock()
	defer mu.Unlock()
	if mockEncProc.ID == id {
		return mockEncProc.ID, mockEncProc.Pk, mockEncProc.Params, nil
	}
	return "", nil, "", errors.New("not found")
}

// IDexists mock: returns true if ID matches stored value
func (m *EncProcModel) IDexists(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()
	return mockEncProc.ID == id, nil
}

// InsertAggregation mock: always succeeds (expand as needed)
func (m *EncProcModel) InsertAggregation(id string, ctAggr []byte, sampleSize int) error {
	return nil
}

// GetAggregationsByID mock: returns a single mock aggregation if ID matches
func (m *EncProcModel) GetAggregationsByID(id string) ([]struct {
	ID         string
	CtAggr     []byte
	SampleSize int
	CreatedAt  string
}, error) {
	mu.Lock()
	defer mu.Unlock()
	if mockEncProc.ID == id {
		return []struct {
			ID         string
			CtAggr     []byte
			SampleSize int
			CreatedAt  string
		}{
			{
				ID:         mockEncProc.ID,
				CtAggr:     []byte{0x04, 0x05, 0x06}, // Example ciphertext bytes
				SampleSize: 2,
				CreatedAt:  "2024-01-01T00:00:00Z",
			},
		}, nil
	}
	return nil, errors.New("not found")
}

// DeleteAggregation mock: always succeeds
func (m *EncProcModel) DeleteAggregation(id string) error {
	return nil
}

// InitializeTables mock: no-op
func (m *EncProcModel) InitializeTables() {}
