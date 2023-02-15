package main

import (
	"fmt"
	"os"
	"pacing.go/shared"
	"time"
)

func main() {
	done := make(chan byte)
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case t := <-ticker.C:
				fmt.Println(t.Format(time.DateTime))
			case <-done:
				ticker.Stop()
			}
		}
	}()
	shared.WaitForSignal(func(_ os.Signal) {
		done <- 1
	})
}
