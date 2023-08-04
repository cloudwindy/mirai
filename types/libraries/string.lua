---@meta

--- 以指定的分隔符分割字符串，返回一个包含所有子字符串的表。
---@param sep string 分隔符
---@return string[]
function string.split(s, sep) end

--- 以空格分割字符串，返回一个包含所有子字符串的表。
---@return string[]
function string.fields(s) end

--- 检查字符串是否包含子字符串。
---@param sub string 子字符串
---@return boolean
function string.includes(s, sub) end

--- 检查字符串是否有指定的前缀。
---@param prefix string 前缀
---@return boolean
function string.startswith(s, prefix) end

--- 检查字符串是否有指定的后缀。
---@param suffix string 后缀
---@return boolean
function string.endswith(s, suffix) end

--- 去除字符串两端的子字符串。
---@param sub string 子字符串
---@return string
function string.trim(s, sub) end

--- 去除字符串两端的 Unicode 空格。
---@return string
function string.trimspace(s) end

--- 去除字符串的前缀。
---@param prefix string 前缀
---@return string
function string.trimend(s, prefix) end

--- 去除字符串的后缀。
---@param suffix string 后缀
---@return string
function string.trimend(s, suffix) end
