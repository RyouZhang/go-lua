package glua

import (
	"sync"
)

// #cgo CFLAGS: -I/opt/luajit/include/luajit-2.1
// #cgo LDFLAGS:  -L/opt/luajit/lib -lluajit -ldl -lm
//#include "glua.h"
import "C"

var (
	gs     *scheduler
	gsOnce sync.Once
)

type scheduler struct {
	queue   chan *glTask
	idle    chan *gluaRT
	freeze  chan *gluaRT
	working int
}

func Scheduler() *scheduler {
	gsOnce.Do(func() {
		gs = &scheduler{
			queue:   make(chan *glTask, 128),
			idle:    make(chan *gluaRT, 16),
			freeze:  make(chan *gluaRT, 16),
			working: 0,
		}
		for i := 0; i < 16; i++ {
			gs.idle <- newGLuaRT()
		}
		go gs.loop()
	})
	return gs
}

func (gs *scheduler) loop() {
	for t := range gs.queue {
		if t.pid == 0 {
			rt := <-gs.idle
			t.pid = rt.id
			gs.working++
			go func() {
				defer func() {
					gs.working--
					gs.idle <- rt
				}()
				var (
					res interface{}
					err error
				)
				if t.lt == nil {
					res, err = rt.call(t.scriptPath, t.methodName, t.args...)
				} else {
					res, err = rt.resume(t.lt, t.args...)
				}
				if err == nil {
					t.callback <- res
				} else {
					if err.Error() == "LUA_YIELD" {
						//wait callback
						if t.lt == nil {
							t.lt = res.(*thread)
						}
					} else {
						t.callback <- err
					}
				}
			}()
		}
	}
}

func (gs *scheduler) pushTask(t *glTask) {
	gs.queue <- t
}
