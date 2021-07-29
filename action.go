package glua

import (
	"context"
)

type LuaNextFunc func(context.Context, ...interface{}) (interface{}, error)

type LuaAction struct {
	script     string
	scriptPath string
	entrypoint string
	params     []interface{}
	nextParams []interface{}
	nextFuncs  map[string]LuaNextFunc
}

func NewLuaAction() *LuaAction {
	return &LuaAction{
		params:    make([]interface{}, 0),
		nextFuncs: make(map[string]LuaNextFunc, 0),
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

func (a *LuaAction) Next(methodName string, method LuaNextFunc) *LuaAction {
	a.nextFuncs[methodName] = method
	return a
}

func (a *LuaAction) Execute(ctx context.Context) (interface{}, error) {
	return getScheduler().do(ctx, a)
}
