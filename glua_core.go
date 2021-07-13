package glua

import (
	"container/list"
	"runtime"
	"sync"
)

// #cgo CFLAGS: -I/usr/local/include/luajit-2.1
// #cgo LDFLAGS:  -L/usr/local/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	core     *gLuaCore
	coreOnce sync.Once
)

type gLuaCore struct {
	queue    chan *gLuaContext
	vms      *list.List
	idles    chan *gLuaVM
	waitting []*gLuaContext
	callback map[int64][]*gLuaContext
}

func getCore() *gLuaCore {
	coreOnce.Do(func() {
		count := runtime.NumCPU()
		core = &gLuaCore{
			queue:    make(chan *gLuaContext, 128),
			vms:      list.New(),
			idles:    make(chan *gLuaVM, count),
			waitting: make([]*gLuaContext, 0),
			callback: make(map[int64][]*gLuaContext),
		}
		for i := 0; i < count; i++ {
			vm := newGLuaVM()
			core.vms.PushBack(vm)
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
				c.vms.PushBack(vm)
				c.scheduler()
			}
		}
	}
}

func (c *gLuaCore) scheduler() {
	current := c.vms.Front()
	for {
		if current == nil {
			return
		}

		vm := current.Value.(*gLuaVM)
		target := c.callback[vm.id]
		if len(target) > 0 {
			temp := current.Next()
			c.vms.Remove(current)
			current = temp

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

		if len(c.waitting) == 0 {
			current = current.Next()
			continue
		}

		temp := current.Next()
		c.vms.Remove(current)
		current = temp

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
