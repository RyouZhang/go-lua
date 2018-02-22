package glua

import (
	"io/ioutil"
	"sync"
)

var (
	scripts  map[string]string
	scriptRW sync.RWMutex
)

func init() {
	scripts = make(map[string]string)
}

func refreshScriptCache() {
	scriptRW.Lock()
	defer scriptRW.Unlock()
	scripts = make(map[string]string)
}

func expireScript(filePath string) {
	scriptRW.Lock()
	defer scriptRW.Unlock()
	delete(scripts, filePath)
}

func loadScript(filePath string) (string, error) {
	scriptRW.RLock()
	target, ok := scripts[filePath]
	scriptRW.RUnlock()
	if ok {
		return target, nil
	}

	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	scriptRW.Lock()
	defer scriptRW.Unlock()

	data := string(raw)
	scripts[filePath] = data
	return data, nil
}
