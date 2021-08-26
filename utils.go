package glua

import (
	// "fmt"
	// "strconv"
	"unsafe"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

func generateLuaStateId(vm *C.struct_lua_State) uintptr {
	return uintptr(unsafe.Pointer(vm))
}

func createLuaState() (uintptr, *C.struct_lua_State) {
	vm := C.luaL_newstate()
	C.lua_gc(vm, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(vm)
	C.lua_gc(vm, C.LUA_GCRESTART, 0)
	C.register_go_method(vm)

	if globalOpts.preloadScriptMethod != nil {
		script := globalOpts.preloadScriptMethod()
		C.gluaL_dostring(vm, C.CString(script))
	}

	return generateLuaStateId(vm), vm
}

func createLuaThread(vm *C.struct_lua_State) (uintptr, *C.struct_lua_State) {
	L := C.lua_newthread(vm)
	return generateLuaStateId(L), L
}
