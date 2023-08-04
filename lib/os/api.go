// Package os implements golang package os functionality for lua.
package os

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	lua "github.com/yuin/gopher-lua"
)

var (
	CmdTimeout = 10
)

// System lua os.system(command) return {status=0, stdout="", stderr=""}
func System(L *lua.LState) int {
	command := L.CheckString(1)
	timeout := time.Duration(L.OptInt64(2, int64(CmdTimeout))) * time.Second
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", command)
	case "windows":
		cmd = exec.Command("cmd.exe", "/C", command)
	default:
		L.RaiseError("unsupported os")
	}

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		L.RaiseError("%v", err)
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		go cmd.Process.Kill()
		L.RaiseError("timeout")
	case err := <-done:
		result := L.NewTable()
		L.SetField(result, "stdout", lua.LString(stdout.String()))
		L.SetField(result, "stderr", lua.LString(stderr.String()))
		L.SetField(result, "status", lua.LNumber(-1))

		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					L.SetField(result, "status", lua.LNumber(int64(status.ExitStatus())))
				}
			}
		} else {
			L.SetField(result, "status", lua.LNumber(0))
		}
		L.Push(result)
		return 1
	}
	return 0
}

// Read lua os.read(filepath) reads the file named by filename and returns the contents, returns string
func Read(L *lua.LState) int {
	filename := L.CheckString(1)
	data, err := os.ReadFile(filename)
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LString(data))
	return 1
}

// Write lua os.write(filepath, data) writes data to the file
func Write(L *lua.LState) int {
	filename := L.CheckString(1)
	data := L.CheckString(2)
	err := os.WriteFile(filename, []byte(data), 0o644)
	if err != nil {
		L.RaiseError("%v", err)
	}
	return 0
}

// Mkdir lua os.mkdir()
func Mkdir(L *lua.LState) int {
	err := os.MkdirAll(L.CheckString(1), 0o755)
	if err != nil {
		L.RaiseError("%v", err)
	}
	return 0
}

// Stat lua os.stat(filename) returns table
func Stat(L *lua.LState) int {
	filename := L.CheckString(1)
	stat, err := os.Stat(filename)
	if err != nil {
		L.RaiseError("%v", err)
	}
	result := L.NewTable()
	result.RawSetString(`is_dir`, lua.LBool(stat.IsDir()))
	result.RawSetString(`size`, lua.LNumber(stat.Size()))
	result.RawSetString(`modtime`, lua.LNumber(stat.ModTime().Unix()))
	result.RawSetString(`mode`, lua.LString(stat.Mode().String()))
	L.Push(result)
	return 1
}

func TmpDir(L *lua.LState) int {
	L.Push(lua.LString(os.TempDir()))
	return 1
}

// Hostname lua os.hostname() returns string
func Hostname(L *lua.LState) int {
	hostname, err := os.Hostname()
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LString(hostname))
	return 1
}

// Pagesize lua os.pagesize() return number
func Pagesize(L *lua.LState) int {
	L.Push(lua.LNumber(os.Getpagesize()))
	return 1
}
