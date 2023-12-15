// Package time implements golang package time functionality for lua.
package time

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Now lua time.unix() returns unix timestamp in seconds (float)
func Now(L *lua.LState) int {
	t := float64(time.Now().UnixNano())
	const s = float64(time.Second)
	L.Push(lua.LNumber(t / s))
	return 1
}

// Milli lua time.milli() converts seconds to milliseconds
func Milli(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.Push(lua.LNumber(time.Now().UnixMilli()))
		return 1
	}
	t := float64(L.CheckNumber(1))
	const ms = float64(time.Second / time.Millisecond)
	L.Push(lua.LNumber(t * ms))
	return 1
}

// Micro lua time.micro() converts seconds to microseconds
func Micro(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.Push(lua.LNumber(time.Now().UnixMicro()))
		return 1
	}
	t := float64(L.CheckNumber(1))
	const us = float64(time.Second / time.Microsecond)
	L.Push(lua.LNumber(t * us))
	return 1
}

// Nano lua time.nano() converts seconds to nanoseconds
func Nano(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.Push(lua.LNumber(time.Now().UnixNano()))
		return 1
	}
	t := float64(L.CheckNumber(1))
	const s = float64(time.Second / time.Nanosecond)
	L.Push(lua.LNumber(t * s))
	return 1
}

// Sleep lua time.sleep(number) port of go time.Sleep(int64)
func Sleep(L *lua.LState) int {
	val := float64(L.CheckNumber(1))
	const sec = float64(time.Second)
	time.Sleep(time.Duration(val * sec))
	return 0
}

// Parse lua time.parse(value, layout, ...location) returns number
func Parse(L *lua.LState) int {
	layout, value := L.CheckString(2), L.CheckString(1)
	var (
		err    error
		result time.Time
	)
	if L.GetTop() > 2 {
		location := L.CheckString(3)
		var loc *time.Location
		loc, err = time.LoadLocation(location)
		if err == nil {
			result, err = time.ParseInLocation(layout, value, loc)
		}
	} else {
		result, err = time.Parse(layout, value)
	}
	if err != nil {
		L.RaiseError("%v", err)
	}
	t := float64(result.UTC().UnixNano())
	const sec = float64(time.Second)
	L.Push(lua.LNumber(t / sec))
	return 1
}

// Format lua time.format(unixts, ...layout, ...location) returns string
func Format(L *lua.LState) int {
	tt := float64(L.CheckNumber(1))
	sec := int64(tt)
	nsec := int64((tt - float64(sec)) * 1000000000)
	result := time.Unix(sec, nsec)
	layout := "Mon Jan 2 15:04:05 -0700 MST 2006"
	if L.GetTop() > 1 {
		layout = L.CheckString(2)
	}
	if L.GetTop() < 3 {
		L.Push(lua.LString(result.Format(layout)))
		return 1
	}
	location := L.CheckString(3)
	loc, err := time.LoadLocation(location)
	if err != nil {
		L.RaiseError("%v", err)
	}
	result = result.In(loc)
	L.Push(lua.LString(result.Format(layout)))
	return 1
}
