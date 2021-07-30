package glua

import (
	"sync"
)

type vmPool struct {
	maxVmCount int
	vmCount    int
	pool       sync.Pool
	vmMutex    sync.Mutex
	vmDic      map[int64]*luaVm
}

func newVMPool(maxVmCount int) *vmPool {
	if maxVmCount < 0 {
		maxVmCount = 4
	}
	if maxVmCount > 16 {
		maxVmCount = 16
	}
	return &vmPool{
		maxVmCount: maxVmCount,
		vmDic:      make(map[int64]*luaVm),
		pool: sync.Pool{
			New: func() interface{} {
				return newLuaVm()
			},
		},
	}
}

func (vp *vmPool) accquire() *luaVm {
	vp.vmMutex.Lock()
	defer vp.vmMutex.Unlock()
	vp.vmCount++
	vm := vp.pool.Get().(*luaVm)
	vp.vmDic[vm.stateId] = vm
	return vm
}

func (vp *vmPool) release(vm *luaVm) {
	vp.vmMutex.Lock()
	defer vp.vmMutex.Unlock()
	delete(vp.vmDic, vm.stateId)
	vp.vmCount--
	if vm.needDestory && vm.resumeCount == 0 {
		vm.destory()
		return
	}
	if vp.vmCount > vp.maxVmCount && vm.resumeCount == 0 {
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
