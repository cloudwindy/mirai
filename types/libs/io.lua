---@meta

--- 完整读取一个文件。
---@return string
function io.readfile(filename) end

--- 覆盖写入一个文件。
---@param data string 内容
function io.writefile(filename, data) end

--- 从输入流复制数据到输出流，直到指定长度（或到达 EOF）或者发生错误。返回已复制数据的长度。
---@param n number? 长度
---@return number
function io.copy(dst, src, n) end