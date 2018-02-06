package glua

import (
	"errors"
	"fmt"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type dummy struct {
	key int64
}

func pushToLua(L *C.struct_lua_State, args ...interface{}) {
	for _, arg := range args {
		switch arg.(type) {
		case string:
			C.lua_pushstring(L, C.CString(arg.(string)))
		case float64:
			C.lua_pushnumber(L, C.lua_Number(arg.(float64)))
		case float32:
			C.lua_pushnumber(L, C.lua_Number(arg.(float32)))
		case uint64:
			C.lua_pushnumber(L, C.lua_Number(arg.(uint64)))
		case int64:
			C.lua_pushnumber(L, C.lua_Number(arg.(int64)))
		case uint32:
			C.lua_pushnumber(L, C.lua_Number(arg.(uint32)))
		case int32:
			C.lua_pushnumber(L, C.lua_Number(arg.(int32)))
		case uint16:
			C.lua_pushnumber(L, C.lua_Number(arg.(uint16)))
		case int16:
			C.lua_pushnumber(L, C.lua_Number(arg.(int16)))
		case uint8:
			C.lua_pushnumber(L, C.lua_Number(arg.(uint8)))
		case int8:
			C.lua_pushnumber(L, C.lua_Number(arg.(int8)))
		case uint:
			C.lua_pushnumber(L, C.lua_Number(arg.(uint)))
		case int:
			C.lua_pushnumber(L, C.lua_Number(arg.(int)))
		default:
			{
				//dummy
			}
		}
	}
}

func pullLuaTable(_L *C.struct_lua_State) interface{} {
	keys := make([]interface{}, 0)
	values := make([]interface{}, 0)

	numKeyCount := 0
	var (
		key   interface{}
		value interface{}
	)
	C.lua_pushnil(_L)
	for C.lua_next(_L, -2) != 0 {
		kType := C.lua_type(_L, -2)
		if kType == 4 {
			key = C.GoString(C.glua_tostring(_L, -2))
		} else {
			key = int(C.lua_tointeger(_L, -2))
			numKeyCount = numKeyCount + 1
		}
		vType := C.lua_type(_L, -1)
		switch vType {
		case 0:
			{
				C.glua_pop(_L, 1)
				continue
			}
		case 1:
			{
				temp := C.lua_toboolean(_L, -1)
				if temp == 1 {
					value = true
				} else {
					value = false
				}
			}
		// case 2:
		// 	{
		// 		ptr := C.glua_touserdata(_L, -1)
		// 		target, ok := objMap[int64(*ptr)]
		// 		if ok == false {
		// 			C.glua_pop(_L, 1)
		// 			continue
		// 		}
		// 		value = target.(map[string]interface{})
		// 	}
		case 3:
			{
				value = C.glua_tonumber(_L, -1)
			}
		case 4:
			{
				value = C.GoString(C.glua_tostring(_L, -1))
			}
		case 5:
			{
				value = pullLuaTable(_L)
			}
		}
		keys = append(keys, key)
		values = append(values, value)
		C.glua_pop(_L, 1)
	}
	if numKeyCount == len(keys) {
		return values
	}
	if numKeyCount == 0 {
		result := make(map[string]interface{})
		for index, key := range keys {
			result[key.(string)] = values[index]
		}
		return result
	} else {
		result := make(map[interface{}]interface{})
		for index, key := range keys {
			result[key] = values[index]
		}
		return result
	}
}

func pullFromLua(L *C.struct_lua_State) interface{} {
	vType := C.lua_type(L, -1)
	fmt.Println(vType)
	switch vType {
	case C.LUA_TBOOLEAN:
		{
			res := C.lua_toboolean(L, -1)
			if res == 0 {
				return false
			}
			return true
		}
	case C.LUA_TNUMBER:
		{
			return C.lua_tonumber(L, -1)
		}
	case C.LUA_TSTRING:
		{
			return C.GoString(C.glua_tostring(L, -1))
		}
	case C.LUA_TTABLE:
		{
			return pullLuaTable(L)
		}
	case C.LUA_TLIGHTUSERDATA:
	case C.LUA_TUSERDATA:
		{
			//dummy
		}
	default:
		{
			panic(errors.New(fmt.Sprintf("Unsupport Type %d", vType)))
		}
	}
	return nil
}
