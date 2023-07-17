---@meta

--- 请求会话
---@class Session
---@field [string] any
local Session = {}

--- 枚举会话的所有键。
---@param self Session
---@return string[]
function Session:keys() end

--- 存储会话到数据库。
---@param self Session
---@param expiry number? 过期时间，单位为小时（可选）
function Session:save(expiry) end

--- 销毁会话。
---@param self Session
function Session:destroy() end