package glua

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type luaVm struct {
	stateId      uintptr
	state        *C.struct_lua_State
	scriptMD5Dic map[string]bool
	resumeCount  int
	needDestory  bool
	threadDic    map[uintptr]*C.struct_lua_State
}

func newLuaVm() *luaVm {
	stateId, state := createLuaState()
	return &luaVm{
		stateId:      stateId,
		state:        state,
		resumeCount:  0,
		needDestory:  false,
		scriptMD5Dic: make(map[string]bool),
		threadDic:    make(map[uintptr]*C.struct_lua_State),
	}
}

func (v *luaVm) run(ctx context.Context, luaCtx *luaContext) {
	defer func() {
		C.lua_gc(v.state, C.LUA_GCCOLLECT, 0)
	}()

	threadId, L := createLuaThread(v.state)

	v.threadDic[threadId] = L

	luaCtx.luaStateId = v.stateId
	luaCtx.luaThreadId = threadId

	ret := C.int(C.LUA_OK)

	if len(luaCtx.act.script) > 0 {
		if len(luaCtx.act.entrypoint) > 0 {
			if len(luaCtx.act.scriptMD5) > 0 {
				if _, ok := v.scriptMD5Dic[luaCtx.act.scriptMD5]; !ok {
					v.scriptMD5Dic[luaCtx.act.scriptMD5] = true
					ret = C.gluaL_dostring(L, C.CString(luaCtx.act.script))
				}
			} else {
				scriptMD5 := fmt.Sprintf("%x", md5.Sum([]byte(luaCtx.act.script)))
				if _, ok := v.scriptMD5Dic[scriptMD5]; !ok {
					v.scriptMD5Dic[scriptMD5] = true
					ret = C.gluaL_dostring(L, C.CString(luaCtx.act.script))
				}
			}
		} else {
			ret = C.gluaL_dostring(L, C.CString(luaCtx.act.script))
		}
	} else {
		raw, err := ioutil.ReadFile(luaCtx.act.scriptPath)
		if err != nil {
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
			return
		}
		if len(luaCtx.act.entrypoint) > 0 {
			scriptMD5 := fmt.Sprintf("%x", md5.Sum(raw))
			if _, ok := v.scriptMD5Dic[scriptMD5]; !ok {
				v.scriptMD5Dic[scriptMD5] = true
				ret = C.gluaL_dostring(L, C.CString(string(raw)))
			}
		} else {
			ret = C.gluaL_dostring(L, C.CString(string(raw)))
		}
	}

	if ret == C.LUA_OK && len(luaCtx.act.entrypoint) > 0 {
		C.glua_getglobal(L, C.CString(luaCtx.act.entrypoint))
		pushToLua(L, luaCtx.act.params...)
		ret = C.lua_resume(L, C.int(len(luaCtx.act.params)))
	}

	switch ret {
	case C.LUA_OK:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "success"})
			luaCtx.status = 3
			count := int(C.lua_gettop(L))
			res := make([]interface{}, count)
			for {
				count = int(C.lua_gettop(L))
				if count == 0 {
					break
				}
				res[count-1] = pullFromLua(L, -1)
				C.glua_pop(L, 1)
			}
			if len(res) > 1 {
				luaCtx.callback <- res
			} else {
				luaCtx.callback <- res[0]
			}
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
		}
	case C.LUA_YIELD:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "yield"})
			luaCtx.status = 2
			v.resumeCount++

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
					getScheduler().luaCtxQueue <- luaCtx
				}()
				method, ok := luaCtx.act.funcs[methodName]
				if ok {
					res, err := method(ctx, args...)
					switch res.(type) {
					case []interface{}:
						luaCtx.act.params = append(res.([]interface{}), err)
					default:
						luaCtx.act.params = []interface{}{res, err}
					}
				} else {
					res, err := callExternMethod(ctx, methodName, args...)
					switch res.(type) {
					case []interface{}:
						luaCtx.act.params = append(res.([]interface{}), err)
					default:
						luaCtx.act.params = []interface{}{res, err}
					}
				}
			}()
		}
	default:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "error"})
			luaCtx.status = 3
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(threadId, L)
		}
	}
}

func (v *luaVm) resume(ctx context.Context, luaCtx *luaContext) {
	defer func() {
		C.lua_gc(v.state, C.LUA_GCCOLLECT, 0)
	}()

	v.resumeCount--
	L := v.threadDic[luaCtx.luaThreadId]
	pushToLua(L, luaCtx.act.params...)
	num := C.lua_gettop(L)
	ret := C.lua_resume(L, num)
	switch ret {
	case C.LUA_OK:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "success"})
			luaCtx.status = 3
			count := int(C.lua_gettop(L))
			res := make([]interface{}, count)
			for {
				count = int(C.lua_gettop(L))
				if count == 0 {
					break
				}
				res[count-1] = pullFromLua(L, -1)
				C.glua_pop(L, 1)
			}
			if len(res) > 1 {
				luaCtx.callback <- res
			} else {
				luaCtx.callback <- res[0]
			}
			close(luaCtx.callback)
			v.destoryThread(luaCtx.luaThreadId, L)
		}
	case C.LUA_YIELD:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "yield"})
			v.resumeCount++
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
					getScheduler().luaCtxQueue <- luaCtx
				}()
				method, ok := luaCtx.act.funcs[methodName]
				if ok {
					res, err := method(ctx, args...)
					switch res.(type) {
					case []interface{}:
						luaCtx.act.params = append(res.([]interface{}), err)
					default:
						luaCtx.act.params = []interface{}{res, err}
					}
				} else {
					res, err := callExternMethod(ctx, methodName, args...)
					switch res.(type) {
					case []interface{}:
						luaCtx.act.params = append(res.([]interface{}), err)
					default:
						luaCtx.act.params = []interface{}{res, err}
					}
				}
			}()
		}
	default:
		{
			metricCounter("glua_action_result_total", 1, map[string]string{"type": "error"})
			luaCtx.status = 3
			luaCtx.callback <- errors.New(C.GoString(C.glua_tostring(L, -1)))
			close(luaCtx.callback)
			v.destoryThread(luaCtx.luaThreadId, L)
		}
	}
}

func (v *luaVm) destoryThread(threadId uintptr, L *C.struct_lua_State) {
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
	C.lua_close(v.state)
	v.state = nil
}
