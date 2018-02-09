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
	Scheduler().queue <- t
	for {
		res := <-t.callback
		switch res.(type) {
		case error:
			{
				if res.(error).Error() == "LUA_YIELD" {
					ctx, err := loadYieldContext(t.lt.vm)
					if err != nil {
						return nil, err
					}
					go func() {
						res, err := callMethod(ctx.methodName, ctx.args...)
						if err == nil {
							t.args = []interface{}{res, nil}
						} else {
							t.args = []interface{}{res, err.Error()}
						}
						Scheduler().queue <- t
					}()
				} else {
					return nil, res.(error)
				}
			}
		default:
			{
				t.lt = nil
				return res, nil
			}
		}
	}
}
