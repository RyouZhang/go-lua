package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodDic map[string]func(...interface{}) (interface{}, error)
)

func init() {
	methodDic = make(map[string]func(...interface{}) (interface{}, error))
}

func RegisterExternMethod(methodName string, method func(...interface{}) (interface{}, error)) error {
	_, ok := methodDic[methodName]
	if ok {
		return errors.New("Duplicate Method Name")
	}
	methodDic[methodName] = method
	return nil
}

//export sync_go_method
func sync_go_method(vm *C.struct_lua_State) C.int {
	count := int(C.lua_gettop(vm))
	args := make([]interface{}, count)
	for {
		count = int(C.lua_gettop(vm))
		if count == 0 {
			break
		}
		args[count-1] = pullFromLua(vm, -1)
		C.glua_pop(vm, 1)
	}
	methodName := args[0].(string)
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = make([]interface{}, 0)
	}

	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		C.lua_pushnil(vm)
		C.lua_pushstring(vm, C.CString("Invalid Method Name"))
		return 2
	}
	res, err := tagetMethod(args...)
	if err != nil {
		pushToLua(vm, 0)
		C.lua_pushstring(vm, C.CString(err.Error()))
		return 2
	} else {
		pushToLua(vm, res)
		C.lua_pushnil(vm)
		return 2
	}
}

//export async_go_method
func async_go_method(vm *C.struct_lua_State) C.int {
	count := int(C.lua_gettop(vm))
	args := make([]interface{}, count)
	for {
		count = int(C.lua_gettop(vm))
		if count == 0 {
			break
		}
		args[count-1] = pullFromLua(vm, -1)
		C.glua_pop(vm, 1)
	}
	methodName := args[0].(string)
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = make([]interface{}, 0)
	}

	storeYieldContext(vm, methodName, args...)
	return 0
}

func callExternMethod(methodName string, args ...interface{}) (interface{}, error) {
	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		return nil, errors.New("Invalid Method Name")
	}
	return tagetMethod(args...)
}
