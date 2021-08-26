package glua

type vmPool struct {
	maxVmCount int
	idleVmDic  map[uintptr]*luaVm
	inUseVmDic map[uintptr]*luaVm
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
		inUseVmDic: make(map[uintptr]*luaVm),
		idleVmDic:  make(map[uintptr]*luaVm),
	}
}

func (vp *vmPool) accquire() *luaVm {
	defer func() {
		metricGauge("glua_vm_idle_count", int64(len(vp.idleVmDic)), nil)
		metricGauge("glua_vm_inuse_count", int64(len(vp.inUseVmDic)), nil)
	}()
	// check idle vm
	for _, vm := range vp.idleVmDic {
		delete(vp.idleVmDic, vm.stateId)
		vp.inUseVmDic[vm.stateId] = vm
		return vm
	}
	// create new vm
	if len(vp.inUseVmDic) == vp.maxVmCount {
		return nil
	}
	vm := newLuaVm()
	vp.inUseVmDic[vm.stateId] = vm
	return vm
}

func (vp *vmPool) release(vm *luaVm) {
	defer func() {
		metricGauge("glua_vm_idle_count", int64(len(vp.idleVmDic)), nil)
		metricGauge("glua_vm_inuse_count", int64(len(vp.inUseVmDic)), nil)
	}()
	delete(vp.inUseVmDic, vm.stateId)
	if vm.needDestory && vm.resumeCount == 0 {
		vm.destory()
	} else {
		vp.idleVmDic[vm.stateId] = vm
	}
}

func (vp *vmPool) find(stateId uintptr) *luaVm {
	defer func() {
		metricGauge("glua_vm_idle_count", int64(len(vp.idleVmDic)), nil)
		metricGauge("glua_vm_inuse_count", int64(len(vp.inUseVmDic)), nil)
	}()
	vm, ok := vp.idleVmDic[stateId]
	if ok {
		vp.inUseVmDic[vm.stateId] = vm
		delete(vp.idleVmDic, stateId)
		return vm
	}
	return nil
}
