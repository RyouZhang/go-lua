package glua

import (
	"context"
	"sync"
)

var (
	scheudlerOnce sync.Once
	schuelder     *vmScheduler
)

type luaContext struct {
	ctx         context.Context
	act         *LuaAction
	luaStateId  int64
	luaThreadId int64
	callback    chan interface{}
	status      int //0 wating, 1 running, 2 yield, 3 finish
}

type vmScheduler struct {
	shutdown    chan bool
	resumes     []*luaContext
	resumeQueue chan *luaContext
	waitQueue   chan *luaContext
	vmQueue     chan *luaVm
	vp          *vmPool
}

func getScheduler() *vmScheduler {
	scheudlerOnce.Do(func() {
		schuelder = &vmScheduler{
			shutdown:    make(chan bool),
			resumes:     make([]*luaContext, 0),
			waitQueue:   make(chan *luaContext, 128),
			resumeQueue: make(chan *luaContext, 128),
			vmQueue:     make(chan *luaVm, 64),
			vp:          newVMPool(16),
		}
		go schuelder.loop()
	})
	return schuelder
}

func (s *vmScheduler) loop() {
	for {
		select {
		case <-s.shutdown:
			{
				return
			}
		case vm := <-s.vmQueue:
			{
				var (
					index  int
					luaCtx *luaContext
				)
				for index, _ = range s.resumes {
					luaCtx = s.resumes[index]
					if luaCtx.luaStateId == vm.stateId {
						break
					}
				}
				if luaCtx == nil {
					s.vp.release(vm)
					continue
				}
				switch {
				case len(s.resumes) == 1:
					s.resumes = []*luaContext{}
				case index == len(s.resumes)-1:
					s.resumes = s.resumes[:index-1]
				case index == 0:
					s.resumes = s.resumes[1:]
				default:
					s.resumes = append(s.resumes[:index], s.resumes[index+1:]...)
				}
				go func() {
					defer func() {
						s.vmQueue <- vm
					}()
					vm.resume(luaCtx.ctx, luaCtx)
				}()
			}
		case luaCtx := <-s.waitQueue:
			{
				//select vm
			RETRY:
				vm := s.vp.accquire()
				if vm.needDestory {
					s.vmQueue <- vm
					goto RETRY
				}
				luaCtx.status = 1
				go func() {
					defer func() {
						s.vmQueue <- vm
					}()
					vm.run(luaCtx.ctx, luaCtx)
				}()
			}
		case luaCtx := <-s.resumeQueue:
			{
				vm := s.vp.find(luaCtx.luaStateId)
				if vm == nil {
					s.resumes = append(s.resumes, luaCtx)
					continue
				}
				go func() {
					defer func() {
						s.vmQueue <- vm
					}()
					vm.resume(luaCtx.ctx, luaCtx)
				}()
			}
		}
	}
}

func (s *vmScheduler) do(ctx context.Context, act *LuaAction) (interface{}, error) {
	luaCtx := &luaContext{
		ctx:         ctx,
		act:         act,
		luaStateId:  0,
		luaThreadId: 0,
		callback:    make(chan interface{}, 1),
		status:      0,
	}

	s.waitQueue <- luaCtx

	res := <-luaCtx.callback
	switch res.(type) {
	case error:
		return nil, res.(error)
	default:
		return res, nil
	}
}
