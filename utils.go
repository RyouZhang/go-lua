package glua

import (
	"fmt"
	"strconv"
	"unsafe"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type CreateLuaStateHook func(L *C.struct_lua_State)

func generateLuaStateId(vm *C.struct_lua_State) int64 {
	ptr := unsafe.Pointer(vm)
	key, _ := strconv.ParseInt(fmt.Sprintf("%d", ptr), 10, 64)
	return key
}

func createLuaState() (int64, *C.struct_lua_State) {
	vm := C.luaL_newstate()
	C.lua_gc(vm, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(vm)
	C.lua_gc(vm, C.LUA_GCRESTART, 0)
	C.register_go_method(vm)

	if globalOpts.createStateHook != nil {
		globalOpts.createStateHook(vm)
	}

	return generateLuaStateId(vm), vm
}

func createLuaThread(vm *C.struct_lua_State) (int64, *C.struct_lua_State) {
	L := C.lua_newthread(vm)
	key := generateLuaStateId(L)

	return key, L
}
