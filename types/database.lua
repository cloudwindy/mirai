---@meta

--- 通用数据库接口
---@class odbclib
odbc = {}

--- 新建数据库连接。
---@param driver string 驱动类型
---@param connection_string string 连接字符串
---@param config ODBC_Config? 配置
---@return ODBC_Connection
---@nodiscard
function odbc.open(driver, connection_string, config) end

--- 数据库连接
---@class ODBC_Connection
local Connection = {}

--- 在数据库中查询资源。
---@param query string SQL 语句
---@return ODBC_QueryResult
---@nodiscard
function Connection:query(query) end

--- 在数据库中执行语句。
---@param query string SQL 语句
---@return ODBC_ExecutionResult
function Connection:exec(query) end

--- 准备一条待查询或执行的语句。
---@param query string SQL 语句
---@return ODBC_PreparedStatement
---@nodiscard
function Connection:stmt(query) end

--- 直接向数据库发送命令，不使用事务。例如```PRAGMA journal_mode = OFF;```
---@param query string SQL 语句
---@return ODBC_QueryResult
function Connection:command(query) end

--- 加载 SQL 文件。
---@param filename string SQL 文件名 (不需要 .sql)
---@return ODBC_ExecutionResult
function Connection:loadfile(filename) end

--- 关闭数据库。
function Connection:close() end

--- 语句
---@class ODBC_PreparedStatement
local PreparedStatement = {}

--- 使用指定的参数，在数据库中查询资源。
---@param ... any 参数
---@return ODBC_QueryResult
---@nodiscard
function PreparedStatement:query(...) end

--- 使用指定的参数，在数据库中执行语句。
---@param ... any 参数
---@return ODBC_ExecutionResult
function PreparedStatement:exec(...) end

--- 销毁准备好的语句。
function PreparedStatement:close() end

--- 配置
---@class ODBC_Config
---@field shared boolean 共享连接
---@field max_connections number 最大连接数
---@field read_only boolean 只读模式

--- 查询结果
---@class ODBC_QueryResult
---@field rows table[] 行
---@field columns table[] 列

--- 执行结果
---@class ODBC_ExecutionResult
---@field rows_affected number 影响行数
---@field last_insert_id number 最后一条数据的 id

db = setmetatable({}, { __index = Connection })
