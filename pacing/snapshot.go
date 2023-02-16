package pacing

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"os"
)

type Record struct {
	LineItemID  uuid.UUID `json:"line_item_id"`
	DailyBudget int64     `json:"daily_budget"`
}

const CurrencyUnit int64 = 1_000_000

func LoadSnapshot(path string) ([]*Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(f)
	var res []*Record
	for {
		r := &Record{}
		err = dec.Decode(r)
		if err != nil {
			if err == io.EOF {
				return res, nil
			}
			return nil, err
		}
		res = append(res, r)
	}
}
