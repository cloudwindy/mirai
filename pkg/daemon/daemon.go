package daemon

import (
	"net"
	"os"
	"os/exec"
	"strconv"
)

const (
	flagWorker = "__MIRAI_WORKER"
)

func Fork(wd string, ln *net.TCPListener) (proc *os.Process, err error) {
	ex, err := os.Executable()
	if err != nil {
		return
	}
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	cmd := &exec.Cmd{
		Path:   ex,
		Dir:    wd,
		Args:   args,
		Env:    os.Environ(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if ln != nil {
		// worker process
		var file *os.File
		file, err = ln.File()
		if err != nil {
			return
		}
		cmd.Args[0] = flagWorker
		cmd.ExtraFiles = []*os.File{file}
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	return cmd.Process, nil
}

func IsChild() bool {
	return os.Args[0] == flagWorker
}

func Forked(listen string) (ln *net.TCPListener, err error) {
	var listener net.Listener
	if os.Args[0] == flagWorker {
		listener, err = net.FileListener(os.NewFile(3, ""))
	} else {
		listener, err = net.Listen("tcp", listen)
	}
	return listener.(*net.TCPListener), err
}

func WritePid(path string) error {
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o600)
}

func ReadPid(path string) (pid int, err error) {
	pidfile, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	return strconv.Atoi(string(pidfile))
}
