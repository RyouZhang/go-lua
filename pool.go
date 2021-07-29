package glua

import (
	"sync"
)

var (
	vmPoolOnce sync.Once
	vp         *vmPool
)

type vmPool struct {
	pool    sync.Pool
	vmMutex sync.Mutex
	vmDic   map[int64]*luaVm
}

func getVmPool() *vmPool {
	vmPoolOnce.Do(func() {
		vp = &vmPool{
			vmDic: make(map[int64]*luaVm),
			pool: sync.Pool{
				New: func() interface{} {
					return newLuaVm()
				},
			},
		}
	})
	return vp
}

func (vp *vmPool) accquire() *luaVm {
	vp.vmMutex.Lock()
	defer vp.vmMutex.Unlock()
	vm := vp.pool.Get().(*luaVm)
	vp.vmDic[vm.stateId] = vm
	return vm
}

func (vp *vmPool) release(vm *luaVm) {
	vp.vmMutex.Lock()
	defer vp.vmMutex.Unlock()
	delete(vp.vmDic, vm.stateId)
	if vm.needDestory && vm.resumeCount == 0 {
		vm.destory()
	} else {
		vp.pool.Put(vm)
	}
}

func (vp *vmPool) find(stateId int64) *luaVm {
	vp.vmMutex.Lock()
	defer vp.vmMutex.Unlock()
	_, ok := vp.vmDic[stateId]
	if ok {
		return nil
	}
Find:
	vm := vp.pool.Get().(*luaVm)
	if vm.stateId == stateId {
		vp.vmDic[stateId] = vm
		return vm
	} else {
		vp.pool.Put(vm)
		goto Find
	}
}
