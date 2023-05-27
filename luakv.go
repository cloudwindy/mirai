package main

import lua "github.com/yuin/gopher-lua"

var lkvExports = map[string]lua.LGFunction{
	"create": lkvCreate,
	"exists": lkvExists,
	"keys":   lkvKeys,
	"drop":   lkvDrop,
}

func lkvCheck(L *lua.LState) *Bucket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Bucket); ok {
		return v
	}
	L.ArgError(1, "bucket expected")
	return nil
}

func lkvCreate(L *lua.LState) int {
	bucket := lkvCheck(L)
	_, err := kvCreateBucket(bucket.DB, bucket.Name)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return 0
}

func lkvGet(L *lua.LState) int {
	bucket := lkvCheck(L)
	key := L.CheckString(2)
	res, err := kvGet(bucket.DB, bucket.Name, key)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	if res == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(*res))
	return 1
}

func lkvExists(L *lua.LState) int {
	bucket := lkvCheck(L)
	res, err := kvExists(bucket.DB, bucket.Name)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	if !res {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LFalse)
	}
	return 1
}

func lkvKeys(L *lua.LState) int {
	bucket := lkvCheck(L)
	res, err := kvKeys(bucket.DB, bucket.Name)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	if res == nil {
		L.Push(lua.LNil)
		return 1
	}
	t := L.NewTable()
	for _, v := range res {
		t.Append(lua.LString(v))
	}
	L.Push(t)
	return 1
}

func lkvPut(L *lua.LState) int {
	bucket := lkvCheck(L)
	key := L.CheckString(2)
	value := L.CheckString(3)
	if err := kvPut(bucket.DB, bucket.Name, key, value); err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return 0
}

func lkvDrop(L *lua.LState) int {
	bucket := lkvCheck(L)
	if err := kvDrop(bucket.DB, bucket.Name); err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return 0
}
