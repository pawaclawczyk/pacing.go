package pacing

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"pacing.go/shared"
	"testing"
)

func CreateSnapshot(n int) (string, []*Record, error) {
	f, err := os.CreateTemp("", "snapshot.json")
	if err != nil {
		return "", nil, err
	}
	defer func() {
		shared.PanicIf(f.Close())
	}()
	res := make([]*Record, n)
	enc := json.NewEncoder(f)
	for i := 0; i < n; i++ {
		rec := &Record{
			LineItemID:  uuid.New(),
			DailyBudget: rand.Int63(),
		}
		if err = enc.Encode(rec); err != nil {
			return "", nil, err
		}
		res[i] = rec
	}
	return f.Name(), res, nil
}

func TestLoadSnapshot(t *testing.T) {
	path, recs, err := CreateSnapshot(3)
	assert.NoError(t, err)

	loaded, err := LoadSnapshot(path)
	assert.NoError(t, err)
	assert.Len(t, loaded, 3)
	for _, rec := range recs {
		assert.Contains(t, loaded, rec)
	}
}
