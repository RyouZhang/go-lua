package glua

import (
	"errors"
	"sync"
	"unsafe"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	rw            sync.RWMutex
	luaStateCache map[int64]*C.struct_lua_State
)

func init() {
	luaStateCache = make(map[int64]*C.struct_lua_State)
}

func generateLuaStateId(vm *C.struct_lua_State) int64 {
	return int64(*((*C.int)(unsafe.Pointer(vm))))
}

func createLuaState() (int64, *C.struct_lua_State) {
	vm := C.luaL_newstate()
	C.lua_gc(vm, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(vm)
	C.lua_gc(vm, C.LUA_GCRESTART, 0)
	vmKey := generateLuaStateId(vm)

	rw.Lock()
	defer rw.Unlock()
	luaStateCache[vmKey] = vm

	return vmKey, vm
}

func createLuaThread(vm *C.struct_lua_State) (int64, *C.struct_lua_State) {
	L := C.lua_newthread(vm)
	vmKey := generateLuaStateId(L)

	rw.Lock()
	defer rw.Unlock()
	luaStateCache[vmKey] = vm

	return vmKey, L
}

func findLuaState(vmKey int64) (*C.struct_lua_State, error) {
	rw.RLock()
	defer rw.RUnlock()

	target, ok := luaStateCache[vmKey]
	if ok {
		return target, nil
	} else {
		return nil, errors.New("Invalid Lua Vm Key")
	}
}

func destoryLuaState(vmKey int64) {
	rw.Lock()
	defer rw.Unlock()

	delete(luaStateCache, vmKey)
}
