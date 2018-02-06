package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodDic map[string]func(... interface{}) (interface{}, error)
)

func init() {
	methodDic = make(map[string]func(... interface{}) (interface{}, error))
	registerGoMethod("test_sum", test_sum)
}

//export call_go_method
func call_go_method(vm *C.struct_lua_State) C.int {
	methodName := C.GoString(C.glua_tostring(vm, -2))
	args := pullFromLua(vm, -1)	
	C.glua_pop(vm, -1)	

	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		C.lua_pushnil(vm)
		C.lua_pushstring(vm, C.CString("Invalid Method Name"))
		return 2
	}
	res, err := tagetMethod(args.([]interface{})...)
	if err != nil {
		C.lua_pushnumber(vm, 0)
		C.lua_pushstring(vm, C.CString(err.Error()))			
		return 2
	} else {	
		C.lua_pushnumber(vm, C.lua_Number(res.(int)))		
		C.lua_pushnil(vm)			
		return 2
	}
}

func registerGoMethod(methodName string, method func(... interface{})(interface{}, error)) error {
	_, ok := methodDic[methodName]
	if ok {
		return errors.New("Duplicate Method Name")
	}
	methodDic[methodName] = method
	return nil
}

func test_sum(args... interface{}) (interface{}, error) {
	sum := 0
	for _, arg := range args {
		switch arg.(type) {
		case C.lua_Number:
			{
				sum = sum + int(arg.(C.lua_Number))
			}
		default:
			{
				return nil, errors.New("Invald Arg Type")
			}
		}
	}
	return sum, nil
}