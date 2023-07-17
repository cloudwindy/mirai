---@meta

--- Cookie 操作类
---@class Cookies
---@field [string] string
local Cookies = {}

--- 在客户端上设置 Cookie。
---@param name string
---@param value string
---@param options? CookieOptions
function Cookies:set(name, value, options) end

function Cookies:clear() end

--- Cookie 选项
---@class CookieOptions
local Cookie = {}

--- Indicates the path that must exist in the requested URL for the browser to send the Cookie header.
---
--- The forward slash (/) character is interpreted as a directory separator, and subdirectories are matched as well. For example, for Path=/docs,
--- * the request paths /docs, /docs/, /docs/Web/, and /docs/Web/HTTP will all match.
--- * the request paths /, /docsets, /fr/docs will not match.
---@type string
Cookie.path = ""

--- Defines the host to which the cookie will be sent.
---
--- Only the current domain can be set as the value, or a domain of a higher order, unless it is a public suffix. Setting the domain will make the cookie available to it, as well as to all its subdomains.
---
--- If omitted, this attribute defaults to the host of the current document URL, not including subdomains.
---
--- Contrary to earlier specifications, leading dots in domain names (```.example.com```) are ignored.
---
--- Multiple host/domain values are not allowed, but if a domain is specified, then subdomains are always included.
---@type string?
Cookie.domain = ""

--- If specified, the cookie becomes a **session cookie**. A session finishes when the client shuts down, after which the session cookie is removed.
---
--- **Warning**: Many web browsers have a session restore feature that will save all tabs and restore them the next time the browser is used. Session cookies will also be restored, as if the browser was never closed.
---
--- Max age and expiry is only set when ```session_only``` is false
---@type boolean?
Cookie.session_only = false

--- Indicates the number of seconds until the cookie expires. A zero or negative number will expire the cookie immediately. If both ```expires``` and ```max_age``` are set, ```max_age``` has precedence.
---@type number?
Cookie.max_age = 0

--- Indicates the maximum lifetime of the cookie as an Unix timestamp.
---
--- When an ```expires``` date is set, the deadline is relative to the client the cookie is being set on, not the server.
---@type number?
Cookie.expires = 0

--- Indicates that the cookie is sent to the server only when a request is made with the https: scheme (except on localhost), and therefore, is more resistant to man-in-the-middle attacks.
---@type boolean?
Cookie.secure = false

--- Forbids JavaScript from accessing the cookie, for example, through the ```Document.cookie``` property. Note that a cookie that has been created with ```http_only``` will still be sent with JavaScript-initiated requests, for example, when calling ```XMLHttpRequest.send()``` or ```fetch()```. This mitigates attacks against cross-site scripting (XSS).
---@type boolean?
Cookie.http_only = false

--- Controls whether or not a cookie is sent with cross-site requests, providing some protection against cross-site request forgery attacks (CSRF).
---
--- The possible attribute values are:
---
--- ```Strict```
--- Means that the browser sends the cookie only for same-site requests, that is, requests originating from the same site that set the cookie. If a request originates from a different domain or scheme (even with the same domain), no cookies with the ```same_site = "Strict"``` attribute are sent.
---
--- ```Lax```
--- Means that the cookie is not sent on cross-site requests, such as on requests to load images or frames, but is sent when a user is navigating to the origin site from an external site (for example, when following a link). This is the default behavior if the ```same_site``` attribute is not specified.
---
--- ```None```
--- means that the browser sends the cookie with both cross-site and same-site requests. The ```secure``` attribute must also be set when setting this value, like so ```same_site = "None", secure = true```.
---@type "Strict" | "Lax" | "None"?
Cookie.same_site = "Lax"