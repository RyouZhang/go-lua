package glua

import (
	"errors"
	"unsafe"

	"github.com/RyouZhang/async-go"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	dummyCache *async.KVCache
	yieldCache *async.KVCache
)

func init() {
	dummyCache = async.NewKVCache()
	yieldCache = async.NewKVCache()
}

//lua dummy method
func pushDummy(vm *C.struct_lua_State, obj interface{}) *C.int {
	vmKey := generateLuaStateId(vm)
	ptr := (*C.int)(unsafe.Pointer(&obj))
	dummyId := int64(*ptr)
	dummyCache.Commit(func(data *async.KVData) (interface{}, error) {
		var target map[int64]interface{}

		temp, err := data.Get(vmKey)
		if err != nil {
			target = make(map[int64]interface{})
			data.Set(vmKey, target)
		} else {
			target = temp.(map[int64]interface{})
		}
		target[dummyId] = obj
		return nil, nil
	})
	return ptr
}

func findDummy(vm *C.struct_lua_State, ptr *C.int) (interface{}, error) {
	vmKey := generateLuaStateId(vm)
	dummyId := int64(*ptr)
	return dummyCache.Commit(func(data *async.KVData) (interface{}, error) {
		temp, err := data.Get(vmKey)
		if err != nil {
			return nil, errors.New("Invalid VMKey")
		}
		target := temp.(map[int64]interface{})
		obj, ok := target[dummyId]
		if false == ok {
			return nil, errors.New("Invalid DummyId")
		}
		return obj, nil
	})
}

func cleanDummy(vm *C.struct_lua_State) {
	vmKey := generateLuaStateId(vm)
	dummyCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Del(vmKey)
		return nil, nil
	})
}

//yield method
func storeYieldContext(vm *C.struct_lua_State, methodName string, args ...interface{}) error {
	if vm == nil {
		return errors.New("Invalid Lua State")
	}
	vmKey := generateLuaStateId(vm)
	_, err := yieldCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Set(vmKey, &yieldContext{methodName: methodName, args: args})
		return nil, nil
	})
	return err
}

func loadYieldContext(vm *C.struct_lua_State) (*yieldContext, error) {
	if vm == nil {
		return nil, errors.New("Invalid Lua State")
	}
	vmKey := generateLuaStateId(vm)
	res, err := yieldCache.Commit(func(data *async.KVData) (interface{}, error) {
		res, err := data.Get(vmKey)
		if err == nil {
			data.Del(vmKey)
		}
		return res, err
	})
	return res.(*yieldContext), err
}

type yieldContext struct {
	methodName string
	args       []interface{}
}

type context struct {
	id         int64
	vmId       int64
	threadId   int64
	scriptPath string
	methodName string
	args       []interface{}
	callback   chan interface{}
}

func generateLuaStateId(vm *C.struct_lua_State) int64 {
	return int64(*((*C.int)(unsafe.Pointer(vm))))
}
