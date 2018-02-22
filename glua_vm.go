package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"


type gLuaVM struct {
	vmId	int64
	vm		*C.struct_lua_State
	threadDic map[int64]*gLuaThread
}

func newGLuaVM() *gLuaVM {
	gl := &gLuaVM{
		threadDic:make(map[int64]*GLuaThread),
	}
	gl.vmId, gl.vm = createLuaState()
	return gl
}

func (gl *gLuaVM)destoryThread(t *GLuaThread) {
	t.destory()
	delete(gl.threadDic, t.id)
	var (
		index C.int
		count C.int
	)
	count = C.lua_gettop(gl.vm)
	for index = 1; index <= count; index++ {
		vType := C.lua_type(gl.vm, index)
		if vType == C.LUA_TTHREAD {
			ptr := C.lua_tothread(gl.vm, index)
			if ptr == t.thread {
				C.lua_remove(gl.vm, index)				
				t.thread = nil
				return
			}
		}
	}
}

func (gl *gLuaVM)call(ctx *GLuaContext) (interface{}, error) {
	thread := newGLuaThread(gl.vm)	
	gl.threadDic[thread.id] = thread
	
	res, err := thread.call(ctx.scriptPath, ctx.methodName, ctx.args...)
	if err == nil {
		gl.destoryThread(thread)
		return res, err
	}
	if err.Error() == "LUA_YIELD" {
		ctx.threadId = thread.id
		return nil, err
	} else {
		gl.destoryThread(thread)
		return res, err
	}
}

func (gl *gLuaVM)resume(ctx *GLuaContext) (interface{}, error) {
	thread, ok := gl.threadDic[ctx.threadId]
	if false == ok {
		return nil, errors.New("Invalid Lua Thread")
	}
	res, err := thread.resume(ctx.args...)
	if err != nil && err.Error() == "LUA_YIELD" {
		return nil, err
	} else {
		gl.destoryThread(thread)
		return res, err
	}
}