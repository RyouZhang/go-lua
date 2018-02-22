package glua

import (
	"sync"
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	yieldCache		map[int64]*gLuaYieldContext
	yieldRW			sync.RWMutex			
)

func init() {
	yieldCache = make(map[int64]*yieldContext)
}


type gLuaYieldContext struct {
	methodName string
	args       []interface{}
}

//yield method
func storeYieldContext(vm *C.struct_lua_State, methodName string, args ...interface{}) {
	if vm == nil {
		return errors.New("Invalid Lua State")
	}
	vmKey := generateLuaStateId(vm)

	yieldRW.Lock()
	defer yieldRW.Unlock()
	yieldCache[vmKey] = &gLuaYieldContext{methodName: methodName, args: args}
}

func loadYieldContext(threadId int64) (*gLuaYieldContext, error) {
	if vm == nil {
		return nil, errors.New("Invalid Lua State")
	}
	yieldRW.RLock()
	defer func() {
		delete(yieldCache, threadId)
		yieldRW.RUnlock()
	}()
	target, ok := yieldCache[threadId]
	if false == ok {
		return nil, errors.New("Invalid Yield Contxt")
	}
	return target, nil
}