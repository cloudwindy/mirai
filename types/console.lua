---@meta

--- 控制台
---@class Console
local Console = {}

--- 输出格式化字符串。
---@param format string 格式字符串
---@param ... any 参数
function Console.print(format, ...) end

--- 以错误颜色输出格式化字符串。
---@param format string 格式字符串
---@param ... any 参数
function Console.fail(format, ...) end

--- 以警告颜色输出格式化字符串。
---@param format string 格式字符串
---@param ... any 参数
function Console.warn(format, ...) end

--- 以信息颜色输出格式化字符串。
---@param format string 格式字符串
---@param ... any 参数
function Console.info(format, ...) end

--- 以成功颜色输出格式化字符串。
---@param format string 格式字符串
---@param ... any 参数
function Console.succ(format, ...) end

--- 清空控制台。
function Console.clear() end

cli = setmetatable({}, { __index = Console })
