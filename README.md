# mirai
以 Golang 编写、在 Lua 虚拟机中运行的高性能 HTTP 服务器框架，参考了 express.js 的设计。
An expressjs-like high-performance http server framework written in Golang and Lua.

## Translation
If you're interested in translating README and documents, start an issue!

按理来说 README 应该写英文版，但是我没有时间进行翻译工作，如能帮助将不胜感谢。

## 简介

Mirai 服务器的设计基本参考了 express.js，以请求方法、路径和处理器组成路由，并按顺序执行。
```lua
app:get('/', function(ctx)
  ctx:send('ok')
end)
app:start()
```
在这个例子中，定义了一个方法为```GET```，路径为```/```，处理器为```function(ctx) ctx:send('ok') end```的路由。之后调用了```app:start()```，服务器以非阻塞的方式开始运行，代码结束。（以下省略```app:start()```）

在收到请求后，会从第一个路由或中间件开始尝试匹配，如果匹配则执行，如果不匹配跳过。

当请求匹配时，服务器会调用处理器，并传入与请求的上下文有关的```ctx```。```ctx```中包含了诸多实用的属性与方法，这里调用的方法是```ctx:send()```，该方法可以传入一个 HTTP 状态码和响应体。之后 HTTP 响应会立即被发送，多次调用无效。

有关于```app```和```ctx```的详细信息，请参考[文档](#文档)。

## 中间件
路由与中间件设计可以参考 [express.js 中的路由与中间件](https://expressjs.com/zh-cn/guide/using-middleware.html)。

```lua
-- 中间件在最后会调用 ctx:next() 以继续路由匹配流程。
app:use('/admin/*', function(ctx)
  -- 如果密码等于 abcd1234：
  if ctx.params['password'] == 'abcd1234' then
    -- 保存状态 ok 为 true。
    ctx.state.ok = true
  end
  -- 继续执行路由。
  ctx:next()
end)

app:get('/admin/portal', function(ctx)
  -- 如果状态 ok：
  if ctx.state.ok then
    -- 返回执行成功的信息。
    ctx:send('authorized!')
  end
  -- 由于响应已经发送，无需再次继续匹配，代码结束。
end)
```

## 文档
请安装 [lua-language-server](https://github.com/LuaLS/lua-language-server)，然后将本 Git 仓库下的 types 设置为```workspace.library```。详细信息请参考 [Libraries](https://github.com/LuaLS/lua-language-server/wiki/Libraries)。

## 注意

### 线程安全

由于处理器会运行在与主线程不同的环境，所以在处理器内不能访问由全局环境创建的任何变量。如果一定要访问，请用```app:set()```设置传递。但是，由于不同的线程同时访问某一变量可能引起数据竞争，即使设置了传递值也不能在运行中改变。

```lua
counter = 0
-- 如果不设置传递，在下文中访问 counter 会直接报错。
app:set('counter', counter)
app:get('/counter/add', function(ctx)
  -- 注意！这里试图改变一个由全局环境创建的值，是错误的。
  counter += 1
end)
```
