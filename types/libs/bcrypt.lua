---@meta

---@class bcryptlib
bcrypt = {}

--- 生成密码的 bcrypt 哈希。
---
--- 该方法运行于保护模式，错误信息在第二个返回值中。
---@param password string
---@return string, string
---@nodiscard
function bcrypt.hash(password) end

--- 比较密码和哈希。
---@param hash string
---@param password string
---@return boolean
---@nodiscard
function bcrypt.compare(hash, password) end

return bcrypt