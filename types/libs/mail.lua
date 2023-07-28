---@meta

---@class maillib
local mail = {}

--- 发送邮件。
---@param uri string SMTP 服务器
---@param from string 发件人地址
---@param to string 收件人地址
---@param body string 正文
function mail.send(uri, from, to, body) end

return mail