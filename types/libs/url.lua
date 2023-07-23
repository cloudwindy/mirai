---@meta

---@class uuidlib
url = {}

--- 解析 URL。
---@param input string 要解析的绝对或相对的输入网址。
---@param base string? 如果 input 是相对的，则为要解析的基本网址。如果 input 是绝对的，则将域名和协议替换上去。
---@return url_URL
function url.new(input, base) end

--- 将不符合规范的路径经过格式化转换为标准路径，解析路径中的 . 与 .. 外，还能去掉多余的斜杠。
---@param input string 要格式化的网址
---@return string
function url.normalize(input) end

--- 将一个表转换为 URL 参数。
---@param params table<string, string> 参数表
---@return string
function url.search(params) end

--- 将字符串进行 URI 编码。
---@param input string 要编码的字符串
---@return string
function url.encode(input) end

--- 将 URI 编码后的字符串进行解码。
---@param input string 要解码的字符串
---@return string
function url.decode(input) end

---@class url_URL
local URL = {}

--- 获取和设置网址的片段部分。
---@type string
URL.hash = ""

--- 获取和设置网址的主机部分。
---@type string
URL.host = ""

--- 获取和设置网址的主机名部分。
--- url.host 和 url.hostname 之间的主要区别是 url.hostname 不包括端口。
---@type string
URL.hostname = ""

--- 获取和设置序列化的网址。
---@type string
URL.href = ""

--- 获取网址的源的只读的序列化。
---@type string
URL.origin = ""

--- 获取和设置网址的密码部分。
---@type string
URL.password = ""

--- 获取和设置网址的路径部分。
---@type string
URL.pathname = ""

--- 获取和设置网址的端口部分。
---
--- 设置端口时值可以是一个数字或者字符串。获取时端口是一个数字。
---@type number|string
URL.port = 0

--- 获取和设置网址的协议部分。
---@type string
URL.protocol = ""

--- 获取和设置网址的序列化的查询部分。
---@type string
URL.search = ""

--- 获取表示网址查询参数的对象。
---
--- 该属性是只读的，要替换 URL 的查询参数，请使用 url.search 和 URL.search 设置器。
---@type table<string, string>
URL.search_params = {}

--- 获取和设置网址的用户名部分。
---@type string
URL.username = ""

--- 将网址的 path 和给定的路径连接到一起并返回一个新的 URL 实例。
---@param path string 路径
---@return url_URL
function URL.join(path) end

return url