package glua

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/ugorji/go/codec"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	methodDic map[string]func(...interface{}) (interface{}, error)
	jh        codec.JsonHandle
)

func init() {
	methodDic = make(map[string]func(...interface{}) (interface{}, error))
	registerGoMethod("test_sum", test_sum)

	jh.DecodeOptions.ReaderBufferSize = 128 * 1024 * 1024
	jh.EncodeOptions.WriterBufferSize = 128 * 1024 * 1024
	jh.DecodeOptions.SignedInteger = true
	jh.DecodeOptions.MapType = reflect.TypeOf(map[string]interface{}(nil))

	registerGoMethod("json_decode", json_decode)
	registerGoMethod("get_es_info", get_es_info)
}

//export sync_go_method
func sync_go_method(vm *C.struct_lua_State) C.int {
	count := int(C.lua_gettop(vm))
	methodName := C.GoString(C.glua_tostring(vm, 1))

	args := make([]interface{}, 0)
	for i := 2; i <= count; i++ {
		args = append(args, pullFromLua(vm, i))
	}
	// args := pullFromLua(vm, 2)
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

func registerGoMethod(methodName string, method func(...interface{}) (interface{}, error)) error {
	_, ok := methodDic[methodName]
	if ok {
		return errors.New("Duplicate Method Name")
	}
	methodDic[methodName] = method
	return nil
}

func callMethod(methodName string, args ...interface{}) (interface{}, error) {
	tagetMethod, ok := methodDic[methodName]
	if false == ok {
		return nil, errors.New("Invalid Method Name")
	}
	return tagetMethod(args...)
}

func test_sum(args ...interface{}) (interface{}, error) {
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

func json_decode(args ...interface{}) (interface{}, error) {
	raw := args[0].(string)

	var res map[string]interface{}
	dec := codec.NewDecoderBytes([]byte(raw), &jh)
	err := dec.Decode(&res)
	return res, err
}

func get_es_info(args ...interface{}) (interface{}, error) {
	res, err := http.Get(args[0].(string))
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}
