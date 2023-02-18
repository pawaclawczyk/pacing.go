package pacing

import (
	"pacing.go/dispatcher"
	"time"
)

type Controller struct {
	planned    *PlannedSpend
	spend      *Spend
	dispatcher *dispatcher.Dispatcher
}

func NewController(path string) (*Controller, error) {
	planned := NewPlannedSpend()
	err := planned.Load(path)
	if err != nil {
		return nil, err
	}
	spend := NewSpend()
	dsp, err := dispatcher.NewDispatcher(MakeWorkloadSplitter(planned, spend, time.Now))
	if err != nil {
		return nil, err
	}
	return &Controller{
		planned:    planned,
		spend:      spend,
		dispatcher: dsp,
	}, nil
}

func (c *Controller) Run() error {
	return c.dispatcher.Run()
}

func (c *Controller) Shutdown() {
	c.dispatcher.Shutdown()
}
