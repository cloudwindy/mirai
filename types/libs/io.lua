---@meta

--- 从输入流复制数据到输出流，直到指定长度（或到达 EOF）或者发生错误。返回已复制数据的长度。
---@param dst file* 输出
---@param src file* 输入
---@param n number? 长度
---@return number
function io.copy(dst, src, n) end