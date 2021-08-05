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
	maxVmSize           int
	preloadScriptMethod func() string
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

func (opt *Options) SetPreloadScripeMethod(method func() string) *Options {
	opt.preloadScriptMethod = method
	return opt
}

func GlobalOptions(opts *Options) {
	locker.Lock()
	defer locker.Unlock()
	globalOpts = opts
}
