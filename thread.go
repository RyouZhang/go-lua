package glua

import (
	"errors"

	"github.com/RyouZhang/async-go"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type thread struct {
	vm       *C.struct_lua_State
	dummyDic map[C.int]interface{}
}

func newThread(vm *C.struct_lua_State) *thread {
	_L := C.lua_newthread(vm)
	C.register_go_method(_L)
	return &thread{
		vm:       _L,
		dummyDic: make(map[C.int]interface{}),
	}
}

func (t *thread) destory(vm *C.struct_lua_State) {
	cleanLuaDummy(vm)
	var (
		index C.int
		count C.int
	)
	count = C.lua_gettop(vm)
	for index = 1; index <= count; index++ {
		vType := C.lua_type(vm, index)
		if vType == C.LUA_TTHREAD {
			ptr := C.lua_tothread(vm, index)
			if ptr == t.vm {
				C.lua_remove(vm, index)
				t.vm = nil
				return
			}
		}
	}
}

func (t *thread) call(scriptPath string, methodName string, args ...interface{}) (interface{}, error) {
	_, err := scripts.Commit(func(data *async.KVData) (interface{}, error) {
		target, err := data.Get(scriptPath)
		if err == nil {
			ret := C.gluaL_dostring(t.vm, C.CString(target.(string)))
			if ret == C.LUA_OK {
				return target, nil
			}
			data.Del(scriptPath)
		}
		script, err := loadScript(scriptPath)
		if err != nil {
			return nil, err
		}
		ret := C.gluaL_dostring(t.vm, C.CString(script))
		if ret == C.LUA_OK {
			data.Set(scriptPath, script)
			return script, nil
		} else {
			errStr := C.GoString(C.glua_tostring(t.vm, -1))
			return nil, errors.New(errStr)
		}
	})
	if err != nil {
		return nil, err
	}
	C.glua_getglobal(t.vm, C.CString(methodName))
	pushToLua(t.vm, args...)

	ret := C.lua_resume(t.vm, C.int(len(args)))
	switch ret {
	case C.LUA_OK:
		{
			var (
				res interface{}
				err interface{}
			)
			num := C.lua_gettop(t.vm)
			if num > 1 {
				err = pullFromLua(t.vm, -1)
				C.lua_remove(t.vm, -1)
				res = pullFromLua(t.vm, -1)
			} else {
				res = pullFromLua(t.vm, -1)
			}
			C.glua_pop(t.vm, -1)
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
			temp := C.GoString(C.glua_tostring(t.vm, -1))
			return nil, errors.New(temp)
		}
	}
}

func (t *thread) resume(args ...interface{}) (interface{}, error) {
	pushToLua(t.vm, args...)
	num := C.lua_gettop(t.vm)
	ret := C.lua_resume(t.vm, num)
	switch ret {
	case C.LUA_OK:
		{
			err := pullFromLua(t.vm, -1)
			C.lua_remove(t.vm, -1)
			res := pullFromLua(t.vm, -1)
			C.glua_pop(t.vm, -1)
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
			temp := C.GoString(C.glua_tostring(t.vm, -1))
			return nil, errors.New(temp)
		}
	}
}
