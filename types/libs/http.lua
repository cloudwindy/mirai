---@meta

---@class httplib
http = {}

--- 新建 HTTP 客户端。
---@param config http_Config?
function http.new(config) end

---@class http_Config
local Config = {}

--- 代理服务器，形如```http(s)://<user>:<password>@host:<port>```。
---@type string?
Config.proxy = ""

--- 超时时间，单位为秒。
---@type number?
Config.timeout = 0

--- 不检查证书。（不推荐）
---@type boolean?
Config.insecure_ssl = false

--- User-Agent 头。
---@type string?
Config.user_agent = "gopher-lua"

--- Authorization 头用户名。
---@type string?
Config.basic_auth_user = ""

--- Authorization 头密码。
---@type string?
Config.basic_auth_password = ""

--- 自定义头。
---@type table?
Config.headers = {}

--- 调试模式。
---@type boolean?
Config.debug = false