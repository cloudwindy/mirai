---@meta

---@class jsonlib
json = {}

--- 将一个 Lua 值转换为 JSON 字符串。
---
--- 该方法运行于保护模式，错误信息在第二个返回值中。
---@param value any
---@return string, string
---@nodiscard
function json.encode(value) end

--- 解析 JSON 字符串，并构造字符串描述的 Lua 值。
---
--- 该方法运行于保护模式，错误信息在第二个返回值中。
---@param data string JSON 字符串
---@return any, string
---@nodiscard
function json.decode(data) end

return json