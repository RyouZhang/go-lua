package glua

import (
	"io/ioutil"

	"github.com/RyouZhang/async-go"
)

var (
	scripts *async.KVCache
)

func init() {
	scripts = async.NewKVCache()
}

func refreshScriptCache() {
	scripts.Commit(func(data *async.KVData) (interface{}, error) {
		data.Clean()
		return nil, nil
	})
}

func expireScript(filePath string) {
	scripts.Commit(func(data *async.KVData) (interface{}, error) {
		data.Del(filePath)
		return nil, nil
	})
}

func loadScript(filePath string) (string, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
