package glua

import (
	"context"
	"errors"
	"sync"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodMu  sync.RWMutex
	methodDic map[string]LuaExternFunc
)

type LuaExternFunc func(context.Context, ...interface{}) (interface{}, error)

func init() {
	methodDic = make(map[string]LuaExternFunc)
}

func RegisterExternMethod(methodName string, method LuaExternFunc) error {
	methodMu.Lock()
	defer methodMu.Unlock()
	_, ok := methodDic[methodName]
	if ok {
		return errors.New("Duplicate Method Name")
	}
	methodDic[methodName] = method
	return nil
}

func callExternMethod(ctx context.Context, methodName string, args ...interface{}) (interface{}, error) {
	methodMu.RLock()
	defer methodMu.RUnlock()
	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		return nil, errors.New("Invalid Method Name")
	}
	return tagetMethod(ctx, args...)
}
