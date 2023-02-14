package shared

import (
	"os"
	"os/signal"
	"syscall"
)

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func WaitForSignal(callback func(sig os.Signal)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigs:
		callback(sig)
	}
}

func PtrString(s string) *string {
	return &s
}
