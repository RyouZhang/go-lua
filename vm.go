package glua

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type luaVm struct {
	stateId     int64
	state       *C.struct_lua_State
	resumeCount int
	needDestory bool
	threadDic   map[int64]*C.struct_lua_State
}

func newLuaVm() *luaVm {
	stateId, state := createLuaState()
	return &luaVm{
		stateId:     stateId,
		state:       state,
		resumeCount: 0,
		needDestory: false,
		threadDic:   make(map[int64]*C.struct_lua_State),
	}
}

func (v *luaVm) run(ctx context.Context, luaCtx *luaContext) {
	threadId, L := createLuaThread(v.state)

	v.threadDic[threadId] = L

	luaCtx.luaStateId = v.stateId
	luaCtx.luaThreadId = threadId

	ret := C.int(C.LUA_OK)

	if len(luaCtx.act.script) > 0 {
		ret = C.gluaL_dostring(L, C.CString(luaCtx.act.script))
	} else {
		raw, err := ioutil.ReadFile(luaCtx.act.scriptPath)
		if err != nil {
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
			return
		}
		ret = C.gluaL_dostring(L, C.CString(string(raw)))
	}

	if ret == C.LUA_OK && len(luaCtx.act.entrypoint) > 0 {
		C.glua_getglobal(L, C.CString(luaCtx.act.entrypoint))
		pushToLua(L, luaCtx.act.params...)
		ret = C.lua_resume(L, C.int(len(luaCtx.act.params)))
	}

	switch ret {
	case C.LUA_OK:
		{
			var (
				res interface{}
				err interface{}
			)
			luaCtx.status = 3
			num := C.lua_gettop(L)
			if num > 1 {
				err = pullFromLua(L, -1)
				C.lua_remove(L, -1)
				res = pullFromLua(L, -1)
			} else {
				res = pullFromLua(L, -1)
			}
			C.glua_pop(L, -1)
			if err != nil {
				luaCtx.callback <- errors.New(err.(string))
			} else {
				luaCtx.callback <- res
			}
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
		}
	case C.LUA_YIELD:
		{
			luaCtx.status = 2

			count := int(C.lua_gettop(L))
			args := make([]interface{}, count)
			for {
				count = int(C.lua_gettop(L))
				if count == 0 {
					break
				}
				args[count-1] = pullFromLua(L, -1)
				C.glua_pop(L, 1)
			}

			methodName := args[0].(string)
			if len(args) > 1 {
				args = args[1:]
			} else {
				args = make([]interface{}, 0)
			}

			go func() {
				defer func() {
					if e := recover(); e != nil {
						err, ok := e.(error)
						if !ok {
							err = errors.New(fmt.Sprintf("%v", e))
						}
						luaCtx.act.params = []interface{}{nil, err}
					}
					getScheduler().resumeQueue <- luaCtx
				}()
				method, ok := luaCtx.act.funcs[methodName]
				if ok {
					res, err := method(ctx, args...)
					luaCtx.act.params = []interface{}{res, err}
				} else {
					res, err := callExternMethod(ctx, methodName, args...)
					luaCtx.act.params = []interface{}{res, err}
				}
			}()
		}
	default:
		{
			luaCtx.status = 3
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
		}
	}
}

func (v *luaVm) resume(ctx context.Context, luaCtx *luaContext) {
	L := v.threadDic[luaCtx.luaThreadId]
	pushToLua(L, luaCtx.act.params...)
	num := C.lua_gettop(L)
	ret := C.lua_resume(L, num)
	switch ret {
	case C.LUA_OK:
		{
			luaCtx.status = 3
			err := pullFromLua(L, -1)
			C.lua_remove(L, -1)
			res := pullFromLua(L, -1)
			C.glua_pop(L, -1)
			if err != nil {
				luaCtx.callback <- errors.New(err.(string))
			} else {
				luaCtx.callback <- res
			}
			close(luaCtx.callback)
			v.destoryThread(luaCtx.luaThreadId, L)
		}
	case C.LUA_YIELD:
		{
			luaCtx.status = 2

			count := int(C.lua_gettop(L))
			args := make([]interface{}, count)
			for {
				count = int(C.lua_gettop(L))
				if count == 0 {
					break
				}
				args[count-1] = pullFromLua(L, -1)
				C.glua_pop(L, 1)
			}

			methodName := args[0].(string)
			if len(args) > 1 {
				args = args[1:]
			} else {
				args = make([]interface{}, 0)
			}

			go func() {
				defer func() {
					if e := recover(); e != nil {
						err, ok := e.(error)
						if !ok {
							err = errors.New(fmt.Sprintf("%v", e))
						}
						luaCtx.act.params = []interface{}{nil, err}
					}
					getScheduler().waitQueue <- luaCtx
				}()
				method, ok := luaCtx.act.funcs[methodName]
				if ok {
					res, err := method(ctx, args...)
					luaCtx.act.params = []interface{}{res, err}
				} else {
					res, err := callExternMethod(ctx, methodName, args...)
					luaCtx.act.params = []interface{}{res, err}
				}
			}()
		}
	default:
		{
			luaCtx.status = 3
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(luaCtx.luaThreadId, L)
		}
	}
}

func (v *luaVm) destoryThread(threadId int64, L *C.struct_lua_State) {
	cleanDummy(L)
	delete(v.threadDic, threadId)
	var (
		index C.int
		count C.int
	)
	count = C.lua_gettop(v.state)
	for index = 1; index <= count; index++ {
		vType := C.lua_type(v.state, index)
		if vType == C.LUA_TTHREAD {
			ptr := C.lua_tothread(v.state, index)
			if ptr == L {
				C.lua_remove(v.state, index)
				L = nil
				return
			}
		}
	}
}

func (v *luaVm) destory() {
}
