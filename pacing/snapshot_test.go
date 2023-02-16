package pacing

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadSnapshot(t *testing.T) {
	f, err := os.CreateTemp("", "snapshot.json")
	assert.NoError(t, err)
	enc := json.NewEncoder(f)
	r1 := &Record{
		LineItemID:  uuid.New(),
		DailyBudget: 1234567890,
	}
	r2 := &Record{
		LineItemID:  uuid.New(),
		DailyBudget: 987654321,
	}
	err = enc.Encode(r1)
	assert.NoError(t, err)
	err = enc.Encode(r2)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)

	loaded, err := LoadSnapshot(f.Name())
	assert.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Contains(t, loaded, r1)
	assert.Contains(t, loaded, r2)
}
