package glua

import (
	"context"
)

type LuaAction struct {
	script     string
	scriptPath string
	entrypoint string
	params     []interface{}
	funcs      map[string]LuaExternFunc
}

func NewLuaAction() *LuaAction {
	return &LuaAction{
		params: make([]interface{}, 0),
		funcs:  make(map[string]LuaExternFunc, 0),
	}
}

func (a *LuaAction) WithScript(script string) *LuaAction {
	a.script = script
	return a
}

func (a *LuaAction) WithScriptPath(scriptPath string) *LuaAction {
	a.scriptPath = scriptPath
	return a
}

func (a *LuaAction) WithEntrypoint(entrypoint string) *LuaAction {
	a.entrypoint = entrypoint
	return a
}

func (a *LuaAction) AddParam(param interface{}) *LuaAction {
	a.params = append(a.params, param)
	return a
}

func (a *LuaAction) AddFunc(methodName string, method LuaExternFunc) *LuaAction {
	a.funcs[methodName] = method
	return a
}

func (a *LuaAction) Execute(ctx context.Context) (interface{}, error) {
	return getScheduler().do(ctx, a)
}
