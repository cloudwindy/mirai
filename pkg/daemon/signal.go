package daemon

import (
	"os"
	"os/signal"
)

type SignalListener struct {
	term chan<- os.Signal
	done chan<- bool
}

func Listen(handler func(os.Signal), signals ...os.Signal) SignalListener {
	cterm := make(chan os.Signal, 1)
	cdone := make(chan bool, 1)
	signal.Notify(cterm, signals...)
	go func() {
		for {
			select {
			case sig := <-cterm:
				// terminate
				handler(sig)
			case <-cdone:
				return
			}
		}
	}()
	return SignalListener{term: cterm, done: cdone}
}

func (ln SignalListener) Close() {
	signal.Stop(ln.term)
	ln.done <- true
}

func ExitHandler(_ os.Signal) {
	os.Exit(1)
}

func Kill(pid int, signal os.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(signal)
}
