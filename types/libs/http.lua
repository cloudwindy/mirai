---@meta

---@class httplib
http = {}

--- 发送请求。
---@param verb string 方法
---@param url string 请求 URL
---@param body string|userdata? 请求体，字符串或者 io.open 打开的文件
---@return http_Response
function http.req(verb, url, body) end

--- 新建客户端。
---@param config http_Config?
---@return http_Client
function http.new(config) end

--- 新建请求。
---@param verb string 方法
---@param url string 请求 URL
---@param body string|userdata? 请求体，字符串或者 io.open 打开的文件
---@return http_Request
function http.newreq(verb, url, body) end

---@class http_Config
local Config = {}

--- 代理服务器，形如```http(s)://<user>:<password>@host:<port>```。
---@type string?
Config.proxy = ""

--- 超时时间，单位为秒。
---@type number?
Config.timeout = 0

--- 不检查证书。*（不推荐）*
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

---@class http_Client
local Client = {}

--- 执行请求。
---@param request http_Request
---@return http_Response
function Client:doreq(request) end

--- HTTP 响应
---@class http_Response
local Response = {}

--- 响应代码。
---@type number
Response.code = 0

--- 响应体。
---@type string
Response.body = ""

--- 响应头。
---@type table<string, string>
Response.headers = {}

--- HTTP 请求
---@class http_Request
local Request = {}

--- 设置 Basic Authorization。
---@param username string 用户名
---@param password string 密码
function Request:auth(username, password) end

--- 设置 HTTP 头。
---@param key string
---@param value string
function Request.set(key, value) end

return http
