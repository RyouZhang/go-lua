package glua

import (
	"sync"
	"errors"
	"unsafe"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	dummyCache		map[int64]map[int64]interface{}
	dummyRW			sync.RWMutex			
)

func init() {
	dummyCache = make(map[int64]map[int64]interface{})
}

//lua dummy method
func pushDummy(vm *C.struct_lua_State, obj interface{}) *C.int {		
	vmKey := generateLuaStateId(vm)
	ptr := (*C.int)(unsafe.Pointer(&obj))
	dummyId := int64(*ptr)

	dummyRW.Lock()
	defer dummyRW.Unlock()

	target, ok := dummyCache[vmKey]
	if false == ok {
		target = make(map[int64]interface{})
		dummyCache[vmKey] = target
	}
	target[dummyId] = obj
	
	return ptr
}

func findDummy(vm *C.struct_lua_State, ptr *C.int) (interface{}, error) {
	vmKey := generateLuaStateId(vm)
	dummyId := int64(*ptr)

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