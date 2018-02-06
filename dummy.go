package glua

import (
	"fmt"
	"errors"
	// "reflect"
	"unsafe"

	"github.com/RyouZhang/async-go"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	dummpCache *async.KVCache
)

func init() {
	dummpCache = async.NewKVCache()
}

func generateStateId(vm *C.struct_lua_State) int64 {
	return int64(*((*C.int)(unsafe.Pointer(vm))))
}

func registerLuaDummy(vm *C.struct_lua_State, obj interface{}) *C.int {
	vmKey := generateStateId(vm)
	ptr := (*C.int)(unsafe.Pointer(&obj))
	dummyId := int64(*ptr)
	dummpCache.Commit(func(data *async.KVData) (interface{}, error) {
		var target map[int64]interface{}

		temp, err := data.Get(vmKey)
		if err != nil {
			target = make(map[int64]interface{})
			data.Set(vmKey, target)
		} else {
			target = temp.(map[int64]interface{})
		}
		target[dummyId] = obj
		return nil, nil
	})
	return ptr
}

func findLuaDummy(vm *C.struct_lua_State, ptr *C.int) (interface{}, error) {
	vmKey := generateStateId(vm)
	dummyId := int64(*ptr)
	fmt.Println(dummyId, vmKey)
	return dummpCache.Commit(func(data *async.KVData) (interface{}, error) {
		temp, err := data.Get(vmKey)
		if err != nil {
			return nil, errors.New("Invalid VMKey")
		}
		target := temp.(map[int64]interface{})
		obj, ok := target[dummyId]
		if false == ok {
			return nil, errors.New("Invalid DummyId")
		}
		return obj, nil
	})
}

func cleanLuaDummy(vm *C.struct_lua_State) {
	vmKey := generateStateId(vm)
	dummpCache.Commit(func(data *async.KVData) (interface{}, error) {
		data.Del(vmKey)
		return nil, nil
	})
}
