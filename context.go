package glua

import (
	"sync"
	"errors"
	"unsafe"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	dummyCache    map[int64]map[int64]interface{}
	yieldCache    map[int64]*yieldContext
	luaStateCache map[int64]*C.struct_lua_State
	rw			sync.RWMutex
)

func init() {
	dummyCache = make(map[int64]map[int64]interface{})
	yieldCache = make(map[int64]*yieldContext)
	luaStateCache = make(map[int64]*C.struct_lua_State)
}

//lua dummy method
func pushDummy(vm *C.struct_lua_State, obj interface{}) *C.int {		
	vmKey := generateLuaStateId(vm)
	ptr := (*C.int)(unsafe.Pointer(&obj))
	dummyId := int64(*ptr)

	rw.Lock()
	func() {
		defer rw.Unlock()
		target, ok := dummyCache[vmKey]
		if false == ok {
			target = make(map[int64]interface{})
			dummyCache[vmKey = target]
		}
		target[dummyId] = obj
	}()
	return ptr
}

func findDummy(vm *C.struct_lua_State, ptr *C.int) (interface{}, error) {
	vmKey := generateLuaStateId(vm)
	dummyId := int64(*ptr)

	rw.RLock()
	defer rw.RUnlock()

	target, ok := dummyCache[vmKey]
	if false == ok {
		return nil, errors.New("Invalid VMKey")
	}
	value, ok := target[dummyId]
	if false == ok {
		return nil, errors.New("Invalid DummyId")		
	}
	return value, nil
}

func cleanDummy(vm *C.struct_lua_State) {
	vmKey := generateLuaStateId(vm)

	rw.Lock()
	defer rw.Unlock()
	delete(dummyCache, vmKey)
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

func loadYieldContext(threadId int64) (*yieldContext, error) {
	if vm == nil {
		return nil, errors.New("Invalid Lua State")
	}
	res, err := yieldCache.Commit(func(data *async.KVData) (interface{}, error) {
		res, err := data.Get(threadId)
		if err == nil {
			data.Del(threadId)
		}
		return res, err
	})
	return res.(*yieldContext), err
}

//lua state emthod
func createLuaState() (int64, *C.struct_lua_State) {
	vm := C.luaL_newstate()
	C.lua_gc(vm, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(vm)
	C.lua_gc(vm, C.LUA_GCRESTART, 0)
	vmKey := generateLuaStateId(vm)

	luaStateCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Set(vmKey, vm)
	})
	return vmKey, vm
}

func createLuaThread(vm *C.struct_lua_State) (int64, *C.struct_lua_State) {
	L := C.lua_newthread(vm)
	vmKey := generateLuaStateId(L)
	luaStateCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Set(vmKey, L)
	})
	return vmKey, L
}

func findLuaState(vmKey int64) (*C.struct_lua_State, error) {
	res, err := luaStateCache.Commit(func(data *async.KVData) (interface{}, error) {
		return data.Get(vmKey)
	})
	if err != nil {
		return nil, err
	}
	return res.(*C.struct_lua_State), nil
}

func destoryLuaState(vmKey int64) {
	luaStateCache.Commit(func(data *async.KVData) (interface{}, error) {
		res, err := data.Get(vmKey)
		if err == nil {
			C.lua_close(res.(*C.struct_lua_State))
		}
		data.Del(vmKey)
		return nil, nil
	})
	dummyCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Del(vmKey)
		return nil, nil
	})
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
