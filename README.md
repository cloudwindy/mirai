# mirai
以 Golang 编写、在 Lua 虚拟机中运行的 HTTP 服务器框架，参考了 express.js 的设计。

An expressjs-like http server framework written in Golang and Lua.

## Translation
If you're interested in translating README and documents, start an issue!

按理来说 README 应该写英文版，但是我没有时间进行翻译工作，如能帮助将不胜感谢。

## 简介

Mirai 服务器的设计基本参考了 express.js，以请求方法、路径和处理器组成路由，按先后顺序执行。
```lua
app:get('/', function(ctx)
  ctx:send('ok')
end)
app:start()
```
例子中定义了方法为```GET```，路径为```/```，处理器为```function(ctx) ctx:send('ok') end```的路由。```app:start()```以非阻塞的方式启动服务器。

收到请求后，会从第一个路由或中间件开始尝试匹配。如果请求匹配，服务器调用处理器，并传入与请求的上下文有关的```ctx```。这里用```ctx:send()```来发送 HTTP 响应。

有关于```app```和```ctx```的详细信息，请参考[文档](#文档)。

## 安装

### 编译

要编译 Mirai，请安装以下环境：
* [Golang](https://go.dev/dl/)
* [Taskfile](https://taskfile.dev/installation)

要为 Windows 平台编译，请同时安装：
* [MinGW-w64](https://www.mingw-w64.org/downloads/)

运行以下命令开始编译：
```
task build
```

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
  -- 当下一条路由执行完毕后，会回到这个位置。
  print('请求处理完毕')
end)

app:get('/admin/portal', function(ctx)
  -- 如果状态 ok：
  if ctx.state.ok then
    -- 返回执行成功的信息。
    ctx:send('authorized!')
  end
  -- 由于响应已经发送，无需继续匹配。
end)
```

## 文档
文档是以类型定义的方式呈现的。

要查看文档，请安装 [lua-language-server](https://github.com/LuaLS/lua-language-server)，然后在 [插件管理器] (https://github.com/LuaLS/lua-language-server/wiki/Addons#vs-code-addon-manager)中找到 Mirai Server 并安装。

## 注意

### 线程安全

由于不同的线程同时访问某一变量可能引起数据竞争，传递值不能在运行中改变。

```lua
counter = 0
app:get('/counter/add', function(ctx)
  -- 注意！这里试图改变一个由全局环境创建的值，是错误的。
  counter += 1
end)
```
