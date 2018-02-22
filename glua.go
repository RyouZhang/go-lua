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
	gLuaCore().push(ctx)

Resume:
	res := <- ctx.callback
	switch res.(type) {
	case error:
		{
			if res.(error).Error() == "LUA_YIELD" {
				//todo process yieldcontxt
				goto Resume
			} else {
				return nil, err
			}
		}
	default:
		{
			return res, nil
		}
	}
}
