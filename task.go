package glua

import ()

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

type glTask struct {
	pid        int64
	vm         *C.struct_lua_State
	scriptPath string
	methodName string
	args       []interface{}
	callback   chan interface{}
}
