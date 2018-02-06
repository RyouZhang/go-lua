package glua

import ()

type glTask struct {
	pid        int64
	lt         *thread
	scriptPath string
	methodName string
	args       []interface{}
	callback   chan interface{}
}
