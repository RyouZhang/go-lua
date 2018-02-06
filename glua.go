package glua

import ()

func Call(filePath string, methodName string, args ...interface{}) (interface{}, error) {
	callback := make(chan interface{})
	defer close(callback)
	t := &glTask{
		scriptPath: filePath,
		methodName: methodName,
		args:       args,
		callback:   callback,
	}
	go Scheduler().pushTask(t)

	res := <-callback
	switch res.(type) {
	case error:
		{
			return nil, res.(error)
		}
	default:
		{
			return res, nil
		}
	}
}
