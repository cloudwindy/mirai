---@meta

--- Mirai 服务器
---@class Application : Router
local Application = {}

--- 启动 Mirai 服务器，开始监听新的连接。
function Application:start() end

--- 停止服务器。
---@param timeout number? 超时
function Application:stop(timeout) end

--- 设置传入处理器中的变量。
---
--- 请注意：值必须是一个只读变量，不能在运行的过程中改变。
---@param key string 变量名
---@param value any 值
function Application:set(key, value) end

--- 路由器
---@class Router
local Router = {}

--- 添加一个中间件。
---@param path string? 路径
---@param middleware fun(ctx: Context) 中间件
---@overload fun(self: Application, middleware: fun(ctx: Context))
function Router:use(path, middleware) end

--- 在指定的路径上，为指定的 HTTP 方法添加处理器。
---@param method string 方法
---@param path string 路径
---@param handler fun(ctx: Context) 处理器
function Router:add(method, path, handler) end

--- 在指定的路径上，为所有的 HTTP 方法添加处理器。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:all(path, handler) end

--- 在指定的路径上，为 GET 方法添加处理器。
---
--- GET 方法请求目标资源。这种请求不会改变数据状态。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:get(path, handler) end

--- 在指定的路径上，为 HEAD 方法添加处理器。
---
--- HEAD 方法请求目标资源的响应头，并且这些响应头与用 GET 方法请求得到的一致。服务器不能返回响应体。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:head(path, handler) end

--- 在指定的路径上，为 POST 方法添加处理器。
---
--- POST 方法向目标资源发送数据。这种请求通常会改变数据的状态。
---
--- PUT 和 POST 方法的区别是，PUT 方法是幂等的：调用一次与连续调用多次效果是相同的（即没有副作用），而连续调用多次相同的 POST 方法可能会产生副作用。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:post(path, handler) end

--- 在指定的路径上，为 PUT 方法添加处理器。
---
--- PUT 方法创建新的资源或用请求体替换目标资源。这种请求通常会改变数据的状态。
---
--- PUT 和 POST 方法的区别是，PUT 方法是幂等的：调用一次与连续调用多次效果是相同的（即没有副作用），而连续调用多次相同的 POST 方法可能会产生副作用。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:put(path, handler) end

--- 在指定的路径上，为 DELETE 方法添加处理器。
---
--- DELETE 方法用于删除目标资源。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:delete(path, handler) end

--- 在指定的路径上，为 DELETE 方法添加处理器。
---
--- CONNECT 方法可以开启与目标之间双向沟通的通道。可以用来创建隧道。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:connect(path, handler) end

--- 在指定的路径上，为 OPTIONS 方法添加处理器。
---
--- OPTIONS 方法请求目标资源的允许方法列表。客户端可以用这个方法指定一个 URL，或者用星号（*）来指代整个服务器。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:options(path, handler) end

--- 在指定的路径上，为 TRACE 方法添加处理器。
--- TRACE 方法请求沿着通往目标资源的路径进行信息回环测试。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:trace(path, handler) end

--- 在指定的路径上，为 TRACE 方法添加处理器。
--- PATCH 方法请求对资源进行部分修改。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:patch(path, handler) end

--- 在指定的路径上，为 WebSocket 协议升级添加处理器。
---@param path string 路径
---@param handler fun(ctx: Context)
function Router:upgrade(path, handler) end

app = setmetatable({}, { __index = Application })