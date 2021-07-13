package main

import (
	"fmt"
	"reflect"
	"time"

	glua "github.com/RyouZhang/go-lua"

	"github.com/ugorji/go/codec"
)

var (
	jh codec.JsonHandle
)

func test_sum(args ...interface{}) (interface{}, error) {
	sum := 0
	for _, arg := range args {
		sum = sum + int(arg.(int64))
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

func main() {
	jh.DecodeOptions.ReaderBufferSize = 128 * 1024 * 1024
	jh.EncodeOptions.WriterBufferSize = 128 * 1024 * 1024
	jh.DecodeOptions.SignedInteger = true
	jh.DecodeOptions.MapType = reflect.TypeOf(map[string]interface{}(nil))

	glua.RegisterExternMethod("json_decode", json_decode)
	glua.RegisterExternMethod("test_sum", test_sum)

	fmt.Println(time.Now())
	res, err := glua.Call("script.lua", "async_json_encode", nil)
	fmt.Println(time.Now())
	fmt.Println(res, err)

	fmt.Println(time.Now())
	res, err = glua.Call("script.lua", "test_args", 69)
	fmt.Println(time.Now())
	fmt.Println(res, err)

	s := time.Now()
	res, err = glua.Call("script.lua", "fib", 35)
	fmt.Println(time.Now().Sub(s))
	fmt.Println(res, err)
}
