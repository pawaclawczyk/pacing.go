package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"pacing.go/shared"
	"path"
)

const lineItemsCount = 1_000_000
const snapshotPath = "tmp/snapshot.json"

type record struct {
	LineItemID  uuid.UUID
	DailyBudget int64
}

const currencyUnit int64 = 1_000_000

func randomRecord() *record {
	return &record{
		LineItemID:  uuid.New(),
		DailyBudget: rand.Int63n(100_000 * currencyUnit),
	}
}

func main() {
	err := os.MkdirAll(path.Dir(snapshotPath), os.ModePerm)
	shared.PanicIf(err)
	f, err := os.Create(snapshotPath)
	shared.PanicIf(err)
	defer func() {
		err := f.Close()
		shared.PanicIf(err)
	}()
	encoder := json.NewEncoder(f)
	for i := 0; i < lineItemsCount; i++ {
		err = encoder.Encode(randomRecord())
		shared.PanicIf(err)
	}
}
