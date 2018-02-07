package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type gluaRT struct {
	id int64
	vm *C.struct_lua_State
}

func newGLuaRT() *gluaRT {
	_L := C.luaL_newstate()
	C.lua_gc(_L, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(_L)
	C.lua_gc(_L, C.LUA_GCRESTART, 0)

	gl := &gluaRT{
		id: generateStateId(_L),
		vm: _L,
	}
	return gl
}

func (gl *gluaRT) destory() {
	if gl.vm != nil {
		C.lua_close(gl.vm)
		gl.vm = nil
	}
}

func (gl *gluaRT) call(scriptPath string, methodName string, args ...interface{}) (interface{}, error) {
	t := newThread(gl.vm)
	res, err := t.call(scriptPath, methodName, args...)
	if err == nil {
		t.destory(gl.vm)
		return res, err
	}
	if err.Error() == "LUA_YIELD" {
		return t, err
	} else {
		t.destory(gl.vm)
		return res, err
	}
}

func (gl *gluaRT) resume(t *thread, args ...interface{}) (interface{}, error) {
	if t == nil {
		return nil, errors.New("Invalid Lua Thread")
	}
	res, err := t.resume(args...)
	if err != nil && err.Error() == "LUA_YIELD" {
		return t, err
	} else {
		t.destory(gl.vm)
		return res, err
	}
}
