package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"pacing.go/pacing"
	"pacing.go/shared"
	"path"
)

const lineItemsCount = 1_000_000
const snapshotPath = "tmp/snapshot.json"

func randomRecord() *pacing.Record {
	return &pacing.Record{
		LineItemID:  uuid.New(),
		DailyBudget: rand.Int63n(100_000 * pacing.CurrencyUnit),
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
