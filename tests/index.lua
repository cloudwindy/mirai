local climate = io.read_file('climate.json')
local json = require 'json'
local c, err = json.decode(climate)
if err then
  cli.fail("%v\n", err)
  return
end

app:get('/climate/:country', function(ctx)
  local country = ctx.params['country']
  local temp = {}
  for _, v in ipairs(c) do
    if v.Country == country then
      table.insert(temp, tonumber(v.AverageTemperature))
    end
  end
  local sum = 0
  for _, v in ipairs(temp) do
    sum = sum + v
  end
  local avg = sum / #temp
  ctx:send(tostring(avg))
end)

cli.succ('部署完成\n')