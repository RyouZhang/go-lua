package glua

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
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
	methodName := C.GoString(C.glua_tostring(vm, 1))

	args := make([]interface{}, 0)
	for i := 2; i <= count; i++ {
		args = append(args, pullFromLua(vm, i))
	}
	C.glua_pop(vm, -1)

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
	methodName := C.GoString(C.glua_tostring(vm, 1))

	args := make([]interface{}, 0)
	for i := 2; i <= count; i++ {
		args = append(args, pullFromLua(vm, i))
	}
	C.glua_pop(vm, -1)
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
