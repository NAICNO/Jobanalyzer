package process

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForSignal(signals ...any) {
	stopSignal := make(chan os.Signal, 1)
	for _, x := range signals {
		signal.Notify(stopSignal, x.(syscall.Signal))
	}
	<-stopSignal
}
