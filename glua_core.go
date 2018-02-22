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
	vms      []*gLuaVM
	idles    chan *gLuaVM
	waitting []*gLuaContext
	callback map[int64][]*gLuaContext
}

func getCore() *gLuaCore {
	coreOnce.Do(func() {
		count := runtime.NumCPU()
		core = &gLuaCore{
			queue:    make(chan *gLuaContext, 128),
			vms:      make([]*gLuaVM, 0),
			idles:    make(chan *gLuaVM, count),
			waitting: make([]*gLuaContext, 0),
			callback: make(map[int64][]*gLuaContext),
		}
		for i := 0; i < count; i++ {
			vm := newGLuaVM()
			core.vms = append(core.vms, vm)
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
					c.waitting = append(c.waitting, ctx)
				} else {
					target, ok := c.callback[ctx.vmId]
					if false == ok {
						target = []*gLuaContext{ctx}
					} else {
						target = append(target, ctx)
					}
					c.callback[ctx.vmId] = target
				}
				c.scheduler()
			}
		case vm := <-c.idles:
			{
				c.vms = append(c.vms, vm)
				c.scheduler()
			}
		}
	}
}

func (c *gLuaCore) scheduler() {
	for {
		if len(c.vms) == 0 {
			return
		}

		vm := c.vms[0]
		if len(c.vms) > 1 {
			c.vms = c.vms[1:]
		} else {
			c.vms = []*gLuaVM{}
		}

		target := c.callback[vm.id]
		if len(target) > 0 {
			ctx := target[0]
			if len(target) > 1 {
				c.callback[vm.id] = target[1:]
			} else {
				c.callback[vm.id] = []*gLuaContext{}
			}
			go func() {
				defer func() {
					c.idles <- vm
				}()
				vm.process(ctx)
			}()
			continue
		}
		if len(c.waitting) > 0 {
			ctx := c.waitting[0]
			if len(target) > 1 {
				c.waitting = c.waitting[1:]
			} else {
				c.waitting = []*gLuaContext{}
			}
			go func() {
				defer func() {
					c.idles <- vm
				}()
				vm.process(ctx)
			}()
		}
	}
}
