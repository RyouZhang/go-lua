package glua

type vmPool struct {
	maxVmCount int
	idleVmDic  map[uintptr]*luaVm
	validVmDic map[uintptr]*luaVm
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
		validVmDic: make(map[uintptr]*luaVm),
		idleVmDic:  make(map[uintptr]*luaVm),
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

func (vp *vmPool) find(stateId uintptr) *luaVm {
	vm, ok := vp.idleVmDic[stateId]
	if ok {
		vp.validVmDic[vm.stateId] = vm
		delete(vp.idleVmDic, stateId)
		return vm
	}
	return nil
}
