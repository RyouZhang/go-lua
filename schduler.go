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
	waitings    []*luaContext
	luaCtxQueue chan *luaContext
	vmQueue     chan *luaVm
	vp          *vmPool
}

func getScheduler() *vmScheduler {
	scheudlerOnce.Do(func() {
		schuelder = &vmScheduler{
			shutdown:    make(chan bool),
			resumes:     make([]*luaContext, 0),
			waitings:    make([]*luaContext, 0),
			luaCtxQueue: make(chan *luaContext, 128),
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
				luaCtx := s.pick(vm.stateId)
				if luaCtx == nil {
					s.vp.release(vm)
					continue
				}
				go s.run(vm, luaCtx)
			}
		case luaCtx := <-s.luaCtxQueue:
			{
				switch luaCtx.status {
				case 0:
					{
						vm := s.vp.accquire()
						if vm == nil {
							s.waitings = append(s.waitings, luaCtx)
							continue
						} else {
							luaCtx.status = 1
							go s.run(vm, luaCtx)
						}
					}
				case 2:
					{
						vm := s.vp.find(luaCtx.luaStateId)
						if vm == nil {
							s.resumes = append(s.resumes, luaCtx)
							continue
						}
						go s.run(vm, luaCtx)
					}
				}
			}
		}
	}
}

func (s *vmScheduler) run(vm *luaVm, luaCtx *luaContext) {
	defer func() {
		s.vmQueue <- vm
	}()
	switch luaCtx.status {
	case 2:
		vm.resume(luaCtx.ctx, luaCtx)
	default:
		vm.run(luaCtx.ctx, luaCtx)
	}
}

func (s *vmScheduler) pick(stateId int64) *luaContext {
	var (
		index  int
		luaCtx *luaContext
	)
	// check resume list
	for index, _ = range s.resumes {
		luaCtx = s.resumes[index]
		if luaCtx.luaStateId == stateId {
			break
		}
	}
	if luaCtx != nil {
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
		return luaCtx
	}
	// check waitings list
	if len(s.waitings) == 0 {
		return nil
	}
	luaCtx = s.waitings[0]
	switch {
	case len(s.waitings) == 1:
		s.waitings = []*luaContext{}
	default:
		s.waitings = s.waitings[1:]
	}
	return luaCtx
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

	s.luaCtxQueue <- luaCtx

	res := <-luaCtx.callback
	switch res.(type) {
	case error:
		return nil, res.(error)
	default:
		return res, nil
	}
}
