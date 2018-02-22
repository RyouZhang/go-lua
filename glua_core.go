package glua

import (
	"runtime"
	"sync"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	core     *gLuaCore
	coreOnce sync.Once
)

type gLuaCore struct {
	queue    chan *gLuaContext
	idleVM   chan *gLuaVM
	waitting []*gLuaContext
	callback map[int64][]*gLuaContext
	working  int
}

func getCore() *gLuaCore {
	coreOnce.Do(func() {
		count := runtime.NumCPU()
		core = &gLuaCore{
			queue:    make(chan *gLuaContext, 128),
			idleVM:   make(chan *gLuaVM, count),
			waitting: make([]*gLuaContext, 0),
			callback: make(map[int64][]*gLuaContext),
			working:  0,
		}
		for i := 0; i < count; i++ {
			vm := newGLuaVM()
			core.idle <- vm
		}
		go core.loop()
	})
	return core
}

func (c *gLuaCore) push(ctx *gLuaContext) {
	c.queue <- ctx
}

func (c *gLuaCore) loop() {
	for {
		select {
		case ctx := <-c.queue:
			{
				if ctx.vmId == 0 {
					c.waitting = append(c.normal, ctx)
				} else {
					target, ok := c.callback[ctx.vmId]
					if false == ok {
						target = []*gLuaContext{ctx}
					} else {
						target = append(target, tx)
					}
					c.callback[ctx.vmId] = target
				}
			}
		}
	}
}
