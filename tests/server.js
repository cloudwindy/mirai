const fs = require('fs')
const Koa = require('koa')
const KoaRouter = require('koa-router')

const file = fs.readFileSync('climate.json')
const climate = JSON.parse(file.toString())
const app = new Koa()

const router = new KoaRouter()
router.get('/api/climate/:name', async (ctx, next) => {
  const name = ctx.params['name']
  if (name === '') {
    return await next()
  }
  const temp = []
  for (const elem of climate) {
    if (elem.Country === ctx.params.name) {
      temp.push(Number(elem.AverageTemperature))
    }
  }
  const avg = temp.reduce((p, c) => p + c, 0) / temp.length
  ctx.body = String(avg)
})
app
  .use(router.routes())
  .use(router.allowedMethods())

app.listen('3000', function () {
  console.log('部署完成')
})