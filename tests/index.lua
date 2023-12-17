-- Mirai Project version: 1.4.0
Climate = ''

app:get('/test', function (ctx)
  ctx:send('Good!')
end)

app:get('/api/climate/:country', function(ctx)
  -- Initialize
  if Climate == '' then
		local f = os.read("climate.json")
    local err
		Climate, err = json.decode(f)
		if err then
			cli.fail("%v\n", err)
			return
		end
  end
  local country = ctx.params['country']
  local temp = {}
  for _, v in ipairs(Climate) do
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