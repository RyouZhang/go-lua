package glua

import (
	"sync"
)

var (
	globalOpts *Options
	locker     sync.Mutex
)

func init() {
	globalOpts = NewOptions()
}

type Options struct {
	maxVmSize       int
	createStateHook CreateLuaStateHook
}

func NewOptions() *Options {
	return &Options{
		maxVmSize: 4,
	}
}

func (opt *Options) WithMaxVMSize(maxVmSize int) *Options {
	opt.maxVmSize = maxVmSize
	return opt
}

func (opt *Options) SetCreateLuaStateHook(method CreateLuaStateHook) *Options {
	opt.createStateHook = method
	return opt
}

func GlobaOptions(opts *Options) {
	locker.Lock()
	defer locker.Unlock()
	globalOpts = opts
}
