package pacing

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"pacing.go/dispatcher"
)

type Bidder struct {
	receiver *dispatcher.Receiver
}

func NewBidder() (*Bidder, error) {
	receiver, err := dispatcher.NewReceiver(func(workload string) {
		log.Debug().Msg(fmt.Sprintf("(bidder) received workload: %s", workload))
	})
	if err != nil {
		return nil, err
	}
	return &Bidder{
		receiver: receiver,
	}, err
}

func (b *Bidder) Run() error {
	return b.receiver.Run()
}

func (b *Bidder) Shutdown() error {
	return b.receiver.Shutdown()
}
