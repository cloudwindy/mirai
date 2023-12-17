-- These is a reasonable Mirai Server manifest file.
-- Change them if you need to.
return {
  -- index: your main lua file
  --        '.' looks for index.lua in the current directory
  index = '.',
  -- root: your static files such as css and js
  root = './static',
  -- listen: if you want to go public, listen on "0.0.0.0:80"
  listen = ':80',
  -- api_base: where's your api endpoint
  api_base = '/api',
  -- admin_base: where's your admin endpoint
  admin_base = '/admin',
  -- data_path: databases path
  data_path = './data',
  -- editing: allow remote editing
  --          this is currently too dangerous to use
  editing = false,

  db = {
    -- db.driver: supports mysql, postgres and sqlite3
    driver = 'sqlite3',
    -- db.conn: connection string
    conn = ':memory:',
    -- db.sql_path: your sql files path
    sql_path = './sql',
  },

  limiter = {
    -- limiter.enabled: enable limiter middleware for api_base
    --                  this will be moved to middleware config later
    enabled = false,
    -- limiter.max: maximum requests of a single ip
    max = 100,
    -- limiter.dur: how long to keep each record (in seconds)
    dur = 1,
  }
}
