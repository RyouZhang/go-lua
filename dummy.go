package glua

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type dummy struct {
	key []byte
	val interface{}
}

var (
	dummyCache map[uintptr]map[uintptr]*dummy
	dummyRW    sync.RWMutex
)

func init() {
	dummyCache = make(map[uintptr]map[uintptr]*dummy)
}

// lua dummy method
func pushDummy(vm *C.struct_lua_State, obj interface{}) unsafe.Pointer {
	vmKey := generateLuaStateId(vm)

	val := reflect.ValueOf(obj)
	var (
		realObj interface{}
		dummyId uintptr
	)

	switch val.Kind() {
	case reflect.Pointer:
		{
			realObj = val.Elem().Interface()
		}
	default:
		{
			realObj = obj
		}
	}

	dObj := &dummy{
		key: []byte(fmt.Sprintf("%p", &realObj)),
		val: obj,
	}

	dummyId = uintptr(unsafe.Pointer(&(dObj.key[0])))

	dummyRW.Lock()
	target, ok := dummyCache[vmKey]
	if false == ok {
		target = make(map[uintptr]*dummy)
		target[dummyId] = dObj
		dummyCache[vmKey] = target
	} else {
		target[dummyId] = dObj
	}
	dummyRW.Unlock()

	return unsafe.Pointer(dummyId)
}

func findDummy(vm *C.struct_lua_State, ptr unsafe.Pointer) (interface{}, error) {
	vmKey := generateLuaStateId(vm)
	dummyId := uintptr(ptr)

	dummyRW.RLock()
	defer dummyRW.RUnlock()

	target, ok := dummyCache[vmKey]
	if false == ok {
		return nil, errors.New("Invalid VMKey")
	}
	dObj, ok := target[dummyId]
	if false == ok {
		return nil, errors.New("Invalid DummyId")
	}
	return dObj.val, nil
}

func cleanDummy(vm *C.struct_lua_State) {
	vmKey := generateLuaStateId(vm)

	dummyRW.Lock()
	defer dummyRW.Unlock()
	delete(dummyCache, vmKey)
}
