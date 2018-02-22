package glua

import (
	"errors"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

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