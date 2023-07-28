---@meta

---@class uuidlib
local uuid = {}

--- 生成一个随机 UUID。
---@return string
---@nodiscard
function uuid.new() end

--- 将 UUID 转换为二进制数据。
---@param uuid string UUID
---@return string
---@nodiscard
function uuid.tobytes(uuid) end

--- 将二进制数据转换为 UUID。
---@param bytes string 二进制数据
---@return string
---@nodiscard
function uuid.frombytes(bytes) end

return uuid