package glua

import ()

func Call(filePath string, methodName string, args ...interface{}) (interface{}, error) {
	callback := make(chan interface{})
	defer close(callback)

	ctx := &gLuaContext {
		vmId: 0,
		threadId: 0,
		scriptPath: filePath,
		methodName: methodName,
		args:       args,
		callback:   callback,
	}
	getCore().push(ctx)

Resume:
	res := <- ctx.callback
	switch res.(type) {
	case error:
		{
			if res.(error).Error() == "LUA_YIELD" {
				yctx, err := loadYieldContext(ctx.threadId)
				if err != nil {
					return nil, err
				}
				go func() {
					res, err := callExternMethod(yctx.methodName, yctx.args...)
					if err == nil {
						ctx.args = []interface{}{res, nil}
					} else {
						ctx.args = []interface{}{res, err.Error()}
					}
					getCore().push(ctx)
				}()
				goto Resume
			} else {
				return nil, res.(error)
			}
		}
	default:
		{
			return res, nil
		}
	}
}
