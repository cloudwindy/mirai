---@meta

--- 完整读取一个文件。
---@return string
function os.read(filename) end

--- 覆盖写入一个文件。
---@param data string 内容
function os.write(filename, data) end

--- 获取文件的信息。
---@return os_Stat
function os.stat(filename) end

--- 逐级新建文件夹。
function os.mkdir(filepath) end

--- 获取临时文件夹。
---@return string
function os.tmpdir() end

--- 获取主机名。
---@return string
function os.hostname() end

--- 获取内存页大小。
---@return number
function os.pagesize() end

--- 文件信息
---@class os_Stat
local stat = {}

--- 是一个文件夹
---@type boolean
stat.is_dir = false

--- 大小
---@type number
stat.size = 0

--- 修改时间
---@type number
stat.modtime = 0

--- 文件模式
---@type string
stat.mode = 'drwxr-xr-x'