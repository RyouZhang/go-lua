package glua

import (
	"errors"

	"github.com/RyouZhang/async-go"
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

func (gl *gluaRT) close() {
	if gl.vm != nil {
		C.lua_close(gl.vm)
		gl.vm = nil
	}
}

func (gl *gluaRT) destoryThread(vm *C.struct_lua_State) {
	var (
		index C.int
		count C.int
	)
	count = C.lua_gettop(gl.vm)
	for index = 1; index <= count; index++ {
		vType := C.lua_type(gl.vm, index)
		if vType == C.LUA_TTHREAD {
			ptr := C.lua_tothread(gl.vm, index)
			if ptr == vm {
				C.lua_remove(gl.vm, index)
				return
			}
		}
	}
}

func (gl *gluaRT) call(scriptPath string, methodName string, args ...interface{}) (interface{}, error) {
	vm := C.lua_newthread(gl.vm)

	_, err := scripts.Commit(func(data *async.KVData) (interface{}, error) {
		target, err := data.Get(scriptPath)
		if err == nil {
			ret := C.gluaL_dostring(vm, C.CString(target.(string)))
			if ret == C.LUA_OK {
				return target, nil
			}
			data.Del(scriptPath)
		}
		script, err := loadScript(scriptPath)
		if err != nil {
			return nil, err
		}
		ret := C.gluaL_dostring(vm, C.CString(script))
		if ret == C.LUA_OK {
			data.Set(scriptPath, script)
			return script, nil
		} else {
			errStr := C.GoString(C.glua_tostring(vm, -1))
			return nil, errors.New(errStr)
		}
	})

	if err != nil {
		gl.destoryThread(vm)
		return nil, err
	}

	C.glua_getglobal(vm, C.CString(methodName))
	pushToLua(vm, args...)

	ret := C.lua_resume(vm, C.int(len(args)))
	switch ret {
	case C.LUA_OK:
		{
			res := pullFromLua(vm)
			C.glua_pop(vm, -1)
			gl.destoryThread(vm)
			return res, nil
		}
	case C.LUA_YIELD:
		{
			return vm, errors.New("LUA_YIELD")
		}
	default:
		{
			temp := C.GoString(C.glua_tostring(vm, -1))
			gl.destoryThread(vm)
			return nil, errors.New(temp)
		}
	}
}

func (gl *gluaRT) resume(vm *C.struct_lua_State, args ...interface{}) (interface{}, error) {
	if vm == nil {
		return nil, errors.New("Invalid Lua thread")
	}
	pushToLua(vm, args...)
	ret := C.lua_resume(vm, C.int(len(args)))
	switch ret {
	case C.LUA_OK:
		{
			res := pullFromLua(vm)
			C.glua_pop(vm, -1)
			gl.destoryThread(vm)
			return res, nil
		}
	case C.LUA_YIELD:
		{
			return vm, errors.New("LUA_YIELD")
		}
	default:
		{
			temp := C.GoString(C.glua_tostring(vm, -1))
			gl.destoryThread(vm)
			return nil, errors.New(temp)
		}
	}
}
