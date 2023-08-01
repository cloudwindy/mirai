---@meta

--- 环境变量
---@class env
---@field [string] any
env = {}

--- Lua 文件根目录
---@type string
env.INDEX = ""

--- HTTP 静态文件根目录
---@type string
env.ROOT = ""

--- 监听地址
---@type string
env.LISTEN = ""

--- Mirai 服务器版本
---@type string
env.VERSION = ""

--- Golang 版本
---@type string
env.GO_VERSION = ""

--- Fiber 版本
---@type string
env.FIBER_VERSION = ""