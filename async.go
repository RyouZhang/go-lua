package glua

import (
	"github.com/RyouZhang/async-go"
)

var (
	contextCache *async.KVCache
)

func init() {
	contextCache = async.NewKVCache()
}

func StoreAsyncContext(vmKey int64, methodName string, args ...interface{}) {
	contextCache.Commit(func(data *async.KVData) (interface{}, error) {
		value := []interface{}{methodName, args}
		data.Set(vmKey, value)
		return nil, nil
	})
}

func LoadAsyncContext(vmKey int64) (string, []interface{}, error) {
	res, err := contextCache.Commit(func(data *async.KVData) (interface{}, error) {
		res, err := data.Get(vmKey)
		if err == nil {
			data.Del(vmKey)
		}
		return res, err
	})
	args := res.([]interface{})
	return args[0].(string), args[1].([]interface{}), err
}
