# Mirai Lua Libraries

## Definition
Libraries are minimal, independent and easy to load on ```lua.LState``` creation. They should not communicate or maintain state across different ```lua.LState```. If you have to, make a ```lue.Engine``` plugin.

## Dependency
Every library in /lib are independent, which means they must not depend on any code in this project outside /lib.

Libraries can depend on each other, but it is not encouraged to do so.

## Coding
1. For ease of use, code should not return error value as in golang. Instead, use ```lua.ArgError``` or ```lua.RaiseError```.
2. Avoid declaring global variables. Use prefixes (eg. ```_pkgname_GlobalVariable```) if you have to declare one.
3. Keep in mind that every time your library is loaded, it can be in different ```lua.LState``` and in different goroutine.

## Copyright
For `http` `io` `odbc` `os` `re` `strings`:

See [original author](https://github.com/vadv/gopher-lua-libs).