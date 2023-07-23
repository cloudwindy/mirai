// Package regexp implements golang package regexp functionality for lua.
package re

import (
	"regexp"

	lua "github.com/yuin/gopher-lua"
)

type luaRegexp struct {
	*regexp.Regexp
}

func checkRegexp(L *lua.LState, n int) *luaRegexp {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*luaRegexp); ok {
		return v
	}
	L.ArgError(n, "regexp_ud expected")
	return nil
}

// regexp.compile(string) returns (regexp_ud, error)
func Compile(L *lua.LState) int {
	expr := L.CheckString(1)
	reg, err := regexp.Compile(expr)
	if err != nil {
		L.RaiseError("%v", err)
	}
	ud := L.NewUserData()
	ud.Value = &luaRegexp{Regexp: reg}
	L.SetMetatable(ud, L.GetTypeMetatable(`regexp_ud`))
	L.Push(ud)
	return 1
}

// regexp_ud:match(string) returns bool
func CompiledMatch(L *lua.LState) int {
	reg := checkRegexp(L, 1)
	str := L.CheckString(2)
	L.Push(lua.LBool(reg.MatchString(str)))
	return 1
}

// regexp.match(regular expression string, string) returns bool
func Match(L *lua.LState) int {
	expr := L.CheckString(1)
	str := L.CheckString(2)
	reg, err := regexp.Compile(expr)
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LBool(reg.MatchString(str)))
	return 1
}

// regexp_ud:findall(string) returns table of table of strings
func CompiledFindAll(L *lua.LState) int {
	reg := checkRegexp(L, 1)
	str := L.CheckString(2)
	result := L.NewTable()
	for _, t := range reg.FindAllStringSubmatch(str, -1) {
		row := L.NewTable()
		for _, v := range t {
			row.Append(lua.LString(v))
		}
		result.Append(row)
	}
	L.Push(result)
	return 1
}

// regexp.findall(regular expression string, string) returns table of table strings
func FindAll(L *lua.LState) int {
	expr := L.CheckString(1)
	str := L.CheckString(2)
	reg, err := regexp.Compile(expr)
	if err != nil {
		L.RaiseError("%v", err)
	}
	result := L.NewTable()
	for _, t := range reg.FindAllStringSubmatch(str, -1) {
		row := L.NewTable()
		for _, v := range t {
			row.Append(lua.LString(v))
		}
		result.Append(row)
	}
	L.Push(result)
	return 1
}
