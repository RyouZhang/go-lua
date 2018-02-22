package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type gLuaVM struct {
	id    int64
	queue chan *gLuaContext
	vm    *C.struct_lua_State
}

func newGLuaVM() *gLuaVM {
	gl := &gLuaVM{
		queue: make(chan *gLuaContext, 128),
	}
	gl.id, gl.vm = createLuaState()
	go gl.loop()
	return gl
}

func (gl *gLuaVM) destory() {
	close(gl.queue)
	C.lua_close(gl.vm)
	gl.vm = nil
}

func (gl *gLuaVM) loop() {
	for ctx := range gl.queue {
		if ctx.vmId == 0 {
			ctx.vmId = gl.id
			res, err := gl.call(ctx)
			if err != nil {
				ctx.callback <- err
			} else {
				ctx.callback <- res
			}
		} else {
			res, err := gl.resume(ctx)
			if err != nil {
				ctx.callback <- err
			} else {
				ctx.callback <- res
			}
		}
	}
}

func (gl *gLuaVM) destoryThread(t *gLuaThread) {
	t.destory()
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

func (gl *gLuaVM) call(ctx *gLuaContext) (interface{}, error) {
	thread := newGLuaThread(gl.vm)
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

func (gl *gLuaVM) resume(ctx *gLuaContext) (interface{}, error) {
	var (
		thread *gLuaThread
		res    interface{}
		err    error
	)

	thread, err = findLuaState(ctx.threadId)
	if err != nil {
		return nil, errors.New("Invalid Lua Thread")
	}
	res, err = thread.resume(ctx.args...)
	if err != nil && err.Error() == "LUA_YIELD" {
		return nil, err
	} else {
		gl.destoryThread(thread)
		return res, err
	}
}
