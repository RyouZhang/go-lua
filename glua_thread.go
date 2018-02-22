package glua

import (
	"errors"

	"github.com/RyouZhang/async-go"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type gLuaThread struct {
	id     int64
	thread *C.struct_lua_State
}

func newGLuaThread(vm *C.struct_lua_State) *gLuaThread {
	gl := &gLuaThread{}
	gl.id, gl.thread = createLuaThread(vm)
	return gl
}

func (t *gLuaThread) destory() {
	cleanDummy(t.thread)
}

func (t *gLuaThread) call(scriptPath string, methodName string, args ...interface{}) (interface{}, error) {
	_, err := scripts.Commit(func(data *async.KVData) (interface{}, error) {
		target, err := data.Get(scriptPath)
		if err == nil {
			ret := C.gluaL_dostring(t.thread, C.CString(target.(string)))
			if ret == C.LUA_OK {
				return target, nil
			}
			data.Del(scriptPath)
		}
		script, err := loadScript(scriptPath)
		if err != nil {
			return nil, err
		}
		ret := C.gluaL_dostring(t.thread, C.CString(script))
		if ret == C.LUA_OK {
			data.Set(scriptPath, script)
			return script, nil
		} else {
			errStr := C.GoString(C.glua_tostring(t.thread, -1))
			return nil, errors.New(errStr)
		}
	})
	if err != nil {
		return nil, err
	}
	C.glua_getglobal(t.thread, C.CString(methodName))
	pushToLua(t.thread, args...)

	ret := C.lua_resume(t.thread, C.int(len(args)))
	switch ret {
	case C.LUA_OK:
		{
			var (
				res interface{}
				err interface{}
			)
			num := C.lua_gettop(t.thread)
			if num > 1 {
				err = pullFromLua(t.thread, -1)
				C.lua_remove(t.thread, -1)
				res = pullFromLua(t.thread, -1)
			} else {
				res = pullFromLua(t.thread, -1)
			}
			C.glua_pop(t.thread, -1)
			if err != nil {
				return nil, errors.New(err.(string))
			}
			return res, nil
		}
	case C.LUA_YIELD:
		{
			return nil, errors.New("LUA_YIELD")
		}
	default:
		{
			temp := C.GoString(C.glua_tostring(t.thread, -1))
			return nil, errors.New(temp)
		}
	}
}

func (t *gLuaThread) resume(args ...interface{}) (interface{}, error) {
	pushToLua(t.thread, args...)
	num := C.lua_gettop(t.thread)
	ret := C.lua_resume(t.thread, num)
	switch ret {
	case C.LUA_OK:
		{
			err := pullFromLua(t.thread, -1)
			C.lua_remove(t.thread, -1)
			res := pullFromLua(t.thread, -1)
			C.glua_pop(t.thread, -1)
			if err != nil {
				return nil, errors.New(err.(string))
			}
			return res, nil
		}
	case C.LUA_YIELD:
		{
			return nil, errors.New("LUA_YIELD")
		}
	default:
		{
			temp := C.GoString(C.glua_tostring(t.thread, -1))
			return nil, errors.New(temp)
		}
	}
}
