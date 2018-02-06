package glua

import (
	"fmt"
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodDic map[string]interface{}
)

func init() {
	methodDic = make(map[string]interface{})
}

//export call_go_method
func call_go_method(vm *C.struct_lua_State) C.int {
	methodName := C.GoString(C.glua_tostring(vm, -2))
	args := pullFromLua(vm, -1)	
	C.glua_pop(vm, -1)	
	
	fmt.Println("step1", methodName, C.lua_gettop(vm), args)
	if methodName == "test_sum" {
		res, err := test_sum(args.([]interface{})...)
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
	C.lua_pushnil(vm)
	C.lua_pushstring(vm, C.CString("Invalid Method Name"))
	return 2
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