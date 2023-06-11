package lkv

import (
	"mirai/lib/kv"

	lua "github.com/yuin/gopher-lua"
	bolt "go.etcd.io/bbolt"
)

type Bucket struct {
	DB   *bolt.DB
	Name string
}

var Exports = map[string]lua.LGFunction{
	"create": Create,
	"exists": Exists,
	"keys":   Keys,
	"drop":   Drop,
}

func Check(L *lua.LState) *Bucket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Bucket); ok {
		return v
	}
	L.ArgError(1, "bucket expected")
	return nil
}

func Create(L *lua.LState) int {
	bucket := Check(L)
	_, err := kv.CreateBucket(bucket.DB, bucket.Name)
	if err != nil {
		L.RaiseError("create bucket failed: %v", err)
	}
	return 0
}

func Get(L *lua.LState) int {
	bucket := Check(L)
	key := L.CheckString(2)
	res, err := kv.Get(bucket.DB, bucket.Name, key)
	if err != nil {
		L.RaiseError("get bucket key failed: %v", err)
	}
	if res == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(*res))
	return 1
}

func Exists(L *lua.LState) int {
	bucket := Check(L)
	res, err := kv.Exists(bucket.DB, bucket.Name)
	if err != nil {
		L.RaiseError("get bucket exists failed: %v", err)
	}
	if !res {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LFalse)
	}
	return 1
}

func Keys(L *lua.LState) int {
	bucket := Check(L)
	res, err := kv.Keys(bucket.DB, bucket.Name)
	if err != nil {
		L.RaiseError("get bucket keys failed: %v", err)
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

func Put(L *lua.LState) int {
	bucket := Check(L)
	key := L.CheckString(2)
	if L.Get(3) == lua.LNil {
		if err := kv.Del(bucket.DB, bucket.Name, key); err != nil {
			L.RaiseError("del bucket key failed: %v", err)
		}
		return 0
	}
	value := L.CheckString(3)
	if err := kv.Put(bucket.DB, bucket.Name, key, value); err != nil {
		L.RaiseError("put bucket key failed: %v", err)
	}
	return 0
}

func Drop(L *lua.LState) int {
	bucket := Check(L)
	if err := kv.Drop(bucket.DB, bucket.Name); err != nil {
		L.RaiseError("drop bucket failed: %v", err)
	}
	return 0
}
