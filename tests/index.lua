-- Mirai Project version: 1.3.5
local climate = os.read('climate.json')
local c, err = json.decode(climate)
if err then
  cli.fail('%v\n', err)
  return
end

app:get('/test', function (ctx)
  ctx:send('Good!')
end)

app:get('/api/climate/:country', function(ctx)
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

app:start()
cli.succ('部署完成\n')