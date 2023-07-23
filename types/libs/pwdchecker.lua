---@meta

--- 密码复杂度检查。第二个返回值指定了复杂度不满足要求的原因。
---@param password string 密码
---@param min_entropy number? 最小熵值
---@return boolean, string?
local function pwdchecker(password, min_entropy) end

return pwdchecker