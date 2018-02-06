package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	rdIds chan int64
)

func init() {
	rdIds = make(chan int64, 1)
	rdIds <- 1
}

func generateGLuaRTID() int64 {
	res := <-rdIds
	rdIds <- (res + 1)
	return res
}

type gluaRT struct {
	id int64
	vm *C.struct_lua_State
}

func newGLuaRT() *gluaRT {
	gl := &gluaRT{
		id: generateGLuaRTID(),
	}
	gl.vm = C.luaL_newstate()
	C.lua_gc(gl.vm, C.LUA_GCSTOP, 0)
	C.luaL_openlibs(gl.vm)
	C.lua_gc(gl.vm, C.LUA_GCRESTART, 0)
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
	if err.Error() == "LUA_YIELD" {
		return t, err
	} else {
		t.destory(gl.vm)
		return res, err
	}
}
