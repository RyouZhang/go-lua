package glua

import ()

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"


type gLuaContext struct {
	id         int64
	vmId       int64
	threadId   int64
	scriptPath string
	methodName string
	args       []interface{}
	callback   chan interface{}
}

type gLuaYieldContext struct {
	methodName string
	args       []interface{}
}