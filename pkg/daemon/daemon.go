package daemon

import (
	"net"
	"os"
	"os/signal"
	"syscall"
)

const gracefulRestart = "__MIRAI_GRACEFUL_RESTART"

func Child(listen string) (ln *net.TCPListener, err error) {
	var listener net.Listener
	if os.Getenv(gracefulRestart) != "" {
		_, err = syscall.Setsid()
		if err != nil {
			return
		}
		listener, err = net.FileListener(os.NewFile(3, ""))
	} else {
		listener, err = net.Listen("tcp", listen)
	}
	if err != nil {
		return nil, err
	}
	return listener.(*net.TCPListener), err
}

func Reload(ln *net.TCPListener, wd string) (fork int, err error) {
	file, err := ln.File()
	if err != nil {
		return
	}
	err = os.Setenv(gracefulRestart, "1")
	if err != nil {
		return
	}
	execSpec := &syscall.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), file.Fd()},
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	}
	ex, err := os.Executable()
	if err != nil {
		return
	}
	return syscall.ForkExec(ex, os.Args, execSpec)
}

type StopSignalHandler struct {
	term chan<- os.Signal
	done chan<- bool
}

func StopSignal() StopSignalHandler {
	cterm := make(chan os.Signal, 1)
	cdone := make(chan bool, 1)
	signal.Notify(cterm, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-cterm:
			// terminate
			os.Exit(1)
		case <-cdone:
			return
		}
	}()
	return StopSignalHandler{term: cterm, done: cdone}
}

func (ln StopSignalHandler) Done() {
	signal.Stop(ln.term)
	ln.done <- true
}
