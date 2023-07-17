---@meta

--- 上下文
---
--- 上下文包括了 HTTP 请求和响应的方法。可以用于获取请求的参数、路由参数、标头等。
---
--- 上下文仅在当前请求中有效，请求结束后如需继续使用，请复制。
---@class Context
local Context = {}

--- 请求的唯一标识符
---@type string
Context.id = ""

--- 请求方法
---@type "GET" | "HEAD" | "POST" | "PUT" | "DELETE" | "CONNECT" | "OPTIONS" | "TRACE" | "PATCH" | string
Context.method = ""

--- 请求 URL，包括了请求协议、域名、端口、路径和参数
---@type string
Context.url = ""

--- 请求路径
---@type string
Context.path = ""

--- 请求体
---@type string
Context.body = ""

--- 请求头
---@type table<string, string>
Context.headers = {}

--- 请求路由参数
---@type table<string, string>
Context.params = {}

--- 请求 Cookies
---@type table<string, Cookie?>
Context.cookies = {}

--- 请求参数
---@type table<string, string>
Context.query = {}

--- 请求会话
---@type Session
Context.sess = {}

--- 发送响应。
---@param self Context
---@param status number? 状态码 (可选，默认为 200)
---@param body string | number | table 响应体
---@overload fun(self: Context, body: string | number | table)
function Context:send(status, body) end
