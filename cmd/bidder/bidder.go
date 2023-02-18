package main

import (
	"os"
	"pacing.go/pacing"
	"pacing.go/shared"
)

func main() {
	bidder, err := pacing.NewBidder()
	shared.PanicIf(err)
	shared.PanicIf(bidder.Run())
	defer func() { shared.PanicIf(bidder.Shutdown()) }()
	shared.WaitForSignal(func(sig os.Signal) {})
}
