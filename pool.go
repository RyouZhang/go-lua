package glua

type vmPool struct {
	maxVmCount int
	idleVmDic  map[int64]*luaVm
	validVmDic map[int64]*luaVm
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
		validVmDic: make(map[int64]*luaVm),
		idleVmDic:  make(map[int64]*luaVm),
	}
}

func (vp *vmPool) accquire() *luaVm {
	// check idle vm
	for _, vm := range vp.idleVmDic {
		delete(vp.idleVmDic, vm.stateId)
		vp.validVmDic[vm.stateId] = vm
		return vm
	}
	// create new vm
	if len(vp.validVmDic) == vp.maxVmCount {
		return nil
	}
	vm := newLuaVm()
	vp.validVmDic[vm.stateId] = vm
	return vm
}

func (vp *vmPool) release(vm *luaVm) {
	delete(vp.validVmDic, vm.stateId)
	if vm.needDestory && vm.resumeCount == 0 {
		vm.destory()
	} else {
		vp.idleVmDic[vm.stateId] = vm
	}
}

func (vp *vmPool) find(stateId int64) *luaVm {
	vm, ok := vp.idleVmDic[stateId]
	if ok {
		vp.validVmDic[vm.stateId] = vm
		delete(vp.idleVmDic, stateId)
		return vm
	}
	return nil
}
