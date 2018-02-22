
function fib(n)
    if n == 0 then
        return 0
    elseif n == 1 then
        return 1
    end
    return fib(n-1) + fib(n-2)
end

function test_args(n)
    res, err = sync_go_method('test_sum', 1,2,3,4,5,6,n)
    if err == nil then
        return res, nil
    else
        return n, err
    end
end

function test_pull_table(obj)
    return {a=true, b=123, c='hello luajit', d={e=12, f='good golang'}, e={1,2,3,4,4}, 1, m=obj}, nil
end

function sync_json_encode()
    return sync_go_method('json_decode', '{"a":"ads","b":12,"c":"sadh"}', 'hello world')
end

function async_json_encode()
    return coroutine.yield(async_go_method('json_decode', '{"a":"ads","b":12,"c":"sadh"}', 'hello world'))
end
