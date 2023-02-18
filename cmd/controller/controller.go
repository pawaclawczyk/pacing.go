package main

import (
	"os"
	"pacing.go/pacing"
	"pacing.go/shared"
)

func main() {
	srv, err := pacing.NewController("tmp/snapshot.json")
	shared.PanicIf(err)
	shared.PanicIf(srv.Run())
	defer srv.Shutdown()
	shared.WaitForSignal(func(sig os.Signal) {})
}
