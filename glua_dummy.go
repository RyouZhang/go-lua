package glua

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"unsafe"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	dummyCache map[int64]map[int64]interface{}
	dummyRW    sync.RWMutex
)

func init() {
	dummyCache = make(map[int64]map[int64]interface{})
}

//lua dummy method
func pushDummy(vm *C.struct_lua_State, obj interface{}) unsafe.Pointer {
	vmKey := generateLuaStateId(vm)

	// addr, _ := strconv.ParseInt(fmt.Sprintf("%d", &obj), 10, 64)
	ptr := unsafe.Pointer(&obj)
	dummyId, _ := strconv.ParseInt(fmt.Sprintf("%d", ptr), 10, 64)

	dummyRW.Lock()
	defer dummyRW.Unlock()

	target, ok := dummyCache[vmKey]
	if false == ok {
		target = make(map[int64]interface{})
		target[dummyId] = obj
		dummyCache[vmKey] = target
	} else {
		target[dummyId] = obj
	}

	return ptr
}

func findDummy(vm *C.struct_lua_State, ptr unsafe.Pointer) (interface{}, error) {
	vmKey := generateLuaStateId(vm)
	dummyId, _ := strconv.ParseInt(fmt.Sprintf("%d", ptr), 10, 64)

	dummyRW.RLock()
	defer dummyRW.RUnlock()

	target, ok := dummyCache[vmKey]
	if false == ok {
		return nil, errors.New("Invalid VMKey")
	}
	value, ok := target[dummyId]
	if false == ok {
		return nil, errors.New("Invalid DummyId")
	}
	return value, nil
}

func cleanDummy(vm *C.struct_lua_State) {
	vmKey := generateLuaStateId(vm)

	dummyRW.Lock()
	defer dummyRW.Unlock()
	delete(dummyCache, vmKey)
}
