package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

func LuaNumberToInt64(value interface{}) (int64, error) {
	switch value.(type) {
	case C.lua_Number:
		{
			return int64(value.(C.lua_Number)), nil
		}
	default:
		{
			return 0, errors.New("Invalid Type")
		}
	}
}

func LuaNumberToInt32(value interface{}) (int32, error) {
	switch value.(type) {
	case C.lua_Number:
		{
			return int32(value.(C.lua_Number)), nil
		}
	default:
		{
			return 0, errors.New("Invalid Type")
		}
	}
}

func LuaNumberToInt(value interface{}) (int, error) {
	switch value.(type) {
	case C.lua_Number:
		{
			return int(value.(C.lua_Number)), nil
		}
	default:
		{
			return 0, errors.New("Invalid Type")
		}
	}
}

func LuaNumberToFloat32(value interface{}) (float32, error) {
	switch value.(type) {
	case C.lua_Number:
		{
			return float32(value.(C.lua_Number)), nil
		}
	default:
		{
			return 0.0, errors.New("Invalid Type")
		}
	}
}

func LuaNumberToFloat64(value interface{}) (float64, error) {
	switch value.(type) {
	case C.lua_Number:
		{
			return float64(value.(C.lua_Number)), nil
		}
	default:
		{
			return 0.0, errors.New("Invalid Type")
		}
	}
}
