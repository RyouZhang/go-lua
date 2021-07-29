package glua

import (
	"context"
	"errors"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodDic map[string]func(context.Context, ...interface{}) (interface{}, error)
)

func init() {
	methodDic = make(map[string]func(context.Context, ...interface{}) (interface{}, error))
}

func RegisterExternMethod(methodName string, method func(context.Context, ...interface{}) (interface{}, error)) error {
	_, ok := methodDic[methodName]
	if ok {
		return errors.New("Duplicate Method Name")
	}
	methodDic[methodName] = method
	return nil
}

func callExternMethod(ctx context.Context, methodName string, args ...interface{}) (interface{}, error) {
	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		return nil, errors.New("Invalid Method Name")
	}
	return tagetMethod(ctx, args...)
}
