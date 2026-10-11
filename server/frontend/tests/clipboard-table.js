// 在项目根目录启动前端 preview（127.0.0.1:4175），用 playwright-cli run-code --filename 执行。
async (page) => {
  // 先打开隔离浏览器会话；只替代 Wails 边界，实际组件和交互逻辑保持原样。
  const assert = (condition, message) => { if (!condition) throw new Error(message) }
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', entry => { if (entry.type() === 'error') errors.push(entry.text()) })
  await page.route('**/favicon.ico', route => route.fulfill({status:204,body:''}))
  await page.addInitScript(() => {
    const logs = [
      {level:'INFO', category:'clipboard', contentId:1, preview:'短预览', message:'收到文本事件', timestamp:1800000000001},
      {level:'INFO', category:'clipboard', contentId:2, preview:'图片 · 1 KB', message:'收到图片事件', timestamp:1800000000002},
      {level:'ERROR', category:'clipboard', contentId:3, preview:'不可用文本', message:'读取事件', timestamp:1800000000003},
      {level:'INFO', category:'connection', contentId:0, message:'连接已建立', timestamp:1800000000004}
    ]
    window.clipTest = {reads:[], copies:[], callbacks:{}, reveals:[], applied:[], failClear:false, hold:false, pending:null}
    window.runtime = {
      EventsOnMultiple: (name, callback) => {window.clipTest.callbacks[name]=callback; return () => {}},
      EventsOff: () => {}, WindowMinimise: () => {}, Environment: () => Promise.resolve({platform:'windows'})
    }
    window.go = {main:{App:{
      GetSettings: async () => ({maxLogs:500}),
      ApplySettings: async raw => {window.clipTest.applied.push(JSON.parse(raw))},
      SetContentRevealed: async value => {window.clipTest.reveals.push(value)},
      GetLogs: async () => logs, GetServerStatus: async () => ({isRunning:false,clientCount:0}),
      GetClientStatus: async () => ({isConnected:false,serverUrl:''}), GetServerClients: async () => [],
      GetLocalIPs: async () => ['127.0.0.1'], GetAutostartStatus: async () => ({enabled:false,mode:'server',error:''}),
      GetLogStorePath: async () => 'test/data/logs.json',
      GetLogContent: async (id) => {
        window.clipTest.reads.push(id)
        if (window.clipTest.hold) await new Promise(resolve => {window.clipTest.pending=resolve})
        if (id===1) return {found:true,type:'text',text:'全文第一行\n全文第二行 '+ '长文本'.repeat(300)}
        if (id===2) return {found:true,type:'image',size:68,imageBase64:'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+j4ioAAAAASUVORK5CYII='}
        return {found:false,error:'测试内容不可用'}
      },
      CopyLogContent: async (id) => {window.clipTest.copies.push(id)},
      ClearLogs: async () => {
        if(window.clipTest.failClear) throw new Error('测试保存失败')
        window.clipTest.callbacks['logs:updated']([])
      },
      HideWindow: async () => {}, Quit: async () => {}
    }}}
  })
  await page.goto('http://127.0.0.1:4175/')
  await page.evaluate(() => {
    localStorage.removeItem('np:clipReveal')
    localStorage.removeItem('np:clipSizes')
  })
  await page.reload()
  const table = page.locator('.log-panel')
  await table.waitFor()
  assert((await page.getByRole('columnheader').allTextContents()).join(',') === '级别,分类,剪贴板内容,时间', '表头不符合预期')
  assert(await page.locator('.body-clip, .clip-resizer').count() === 0, '独立面板仍存在')
  assert(await page.locator('.clip-cell').count() === 3, '剪贴板单元格缺失')
  assert(await page.evaluate(() => window.clipTest.reads.length) === 0, '隐藏状态读取了正文')
  assert(await page.evaluate(() => window.clipTest.reveals[0]) === false, '初始隐藏偏好未同步后端')

  const first = page.locator('.clip-cell').nth(0)
  await first.getByRole('button', {name:'显示内容', exact:true}).first().click()
  await first.locator('.clip-text').waitFor()
  assert((await first.locator('.clip-text').innerText()).includes('全文第一行\n全文第二行'), '全文或换行丢失')
  await first.getByRole('button', {name:'复制到剪贴板', exact:true}).click()
  assert(await page.evaluate(() => window.clipTest.copies[0]) === 1, '复制了错误的内容 ID')
  await first.getByRole('button', {name:'隐藏内容', exact:true}).click()
  await first.getByRole('button', {name:'显示内容', exact:true}).first().click()
  assert(await page.evaluate(() => window.clipTest.reads.filter(id => id===1).length) === 1, '重复读取缓存正文')

  const image = page.locator('.clip-cell').nth(1)
  await image.getByRole('button', {name:'显示内容', exact:true}).first().click()
  await image.getByRole('img', {name:'剪贴板图片'}).waitFor()
  await page.waitForFunction(() => document.querySelector('.clip-dim')?.textContent.includes('1 × 1 px'))
  await image.getByRole('img', {name:'剪贴板图片'}).click()
  await page.locator('.n-image-preview').waitFor()
  await page.keyboard.press('Escape')
  await page.locator('.n-image-preview').waitFor({state:'hidden'})

  await page.getByRole('switch', {name:'显示全部剪贴板内容'}).click()
  await page.getByRole('alert').filter({hasText:'测试内容不可用'}).waitFor()
  assert(await page.evaluate(() => localStorage.getItem('np:clipReveal')) === '1', '全局揭示偏好未持久化')
  await page.waitForFunction(() => window.clipTest.reveals.at(-1) === true)

  await table.getByRole('button', {name:'尺寸', exact:true}).click()
  await page.locator('.size-row input').nth(0).fill('720')
  await page.keyboard.press('Tab')
  await page.locator('.size-row input').nth(2).fill('720')
  await page.keyboard.press('Tab')
  await page.waitForFunction(() => JSON.parse(localStorage.getItem('np:clipSizes')).imgW === 720)
  await table.getByRole('button', {name:'尺寸', exact:true}).click()

  for (const theme of ['light', 'dark']) {
    for (const width of [900, 1200, 1500]) {
      await page.setViewportSize({width,height:800})
      await page.evaluate(value => localStorage.setItem('np:colorScheme',value), theme)
      await page.reload()
      await page.locator('.clip-cell img').waitFor()
      const overflow = await page.evaluate(() => {
        const table = document.querySelector('.log-panel')
        return document.documentElement.scrollWidth > innerWidth || table.scrollWidth > table.clientWidth + 1
      })
      assert(!overflow, `${theme}/${width}px 页面或表格溢出`)
      assert((await page.locator('.clip-text').innerText()).includes('全文第二行'), '重载丢失全局显示偏好')
    }
  }

  await table.locator('.n-select').first().click()
  await page.getByText('全部分类', {exact:true}).click()
  await table.getByText('连接已建立', {exact:true}).waitFor()
  await table.locator('.n-select').first().click()
  await page.getByText('剪贴板', {exact:true}).last().click()
  assert(!(await table.innerText()).includes('连接已建立'), '分类筛选未生效')

  // 无原文的旧剪贴板预览仍遵守全局隐藏。
  await page.evaluate(() => window.clipTest.callbacks['logs:updated']([
    {level:'INFO',category:'clipboard',contentId:0,preview:'旧记录预览',message:'旧事件',timestamp:1800000000010}
  ]))
  await page.getByText('旧记录预览',{exact:true}).waitFor()
  await page.getByRole('switch', {name:'显示全部剪贴板内容'}).click()
  assert(!(await table.innerText()).includes('旧记录预览'), '隐藏时暴露了旧记录预览')

  await page.evaluate(() => {
    window.clipTest.hold=true
    window.clipTest.callbacks['logs:updated']([{level:'INFO',category:'clipboard',contentId:1,preview:'短预览',message:'事件',timestamp:1800000000020}])
  })
  await page.locator('.clip-cell').getByRole('button', {name:'显示内容',exact:true}).first().click()
  await page.waitForFunction(() => window.clipTest.pending !== null)
  await table.getByRole('button', {name:'清空',exact:true}).click()
  await page.evaluate(() => window.clipTest.pending())
  assert(await page.locator('.clip-cell').count() === 0, '清空后延迟读取复活了单元格')

  await page.evaluate(() => {
    window.clipTest.failClear=true
    window.clipTest.callbacks['logs:updated']([{level:'INFO',category:'clipboard',contentId:1,preview:'短预览',message:'事件',timestamp:1800000000030}])
  })
  await page.locator('.clip-cell').waitFor()
  await table.getByRole('button', {name:'清空',exact:true}).click()
  await page.getByText('清空记录失败：Error: 测试保存失败',{exact:true}).waitFor()
  assert(await page.locator('.clip-cell').count() === 1, '清空失败时删除了现有内容')
  await page.evaluate(() => window.clipTest.callbacks['settings:open']())
  await page.getByText('系统设置', {exact:true}).waitFor()
  await page.evaluate(() => window.clipTest.callbacks['settings:open']())
  assert(await page.getByText('系统设置', {exact:true}).count() === 1, '重复托盘右键产生多个设置面板')
  assert(await page.getByRole('button', {name:'退出应用',exact:true}).isVisible(), '设置中缺失退出入口')
  await page.evaluate(() => window.clipTest.callbacks['autostart:changed'](true))
  assert(await page.getByRole('checkbox', {name:'随操作系统自动启动'}).isChecked(), '托盘自启事件未更新开关')
  await page.waitForFunction(() => window.clipTest.applied.at(-1)?.autoStartEnabled === true)
  await page.evaluate(() => window.clipTest.callbacks['autostart:changed'](false))
  await page.waitForFunction(() => window.clipTest.applied.at(-1)?.autoStartEnabled === false)
  await page.evaluate(() => window.clipTest.callbacks['window:restored']())
  await page.getByText('系统设置', {exact:true}).waitFor({state:'hidden'})
  await page.evaluate(() => {
    window.clipTest.callbacks['settings:open']()
    window.clipTest.callbacks['record:focus'](1)
  })
  await page.getByText('系统设置', {exact:true}).waitFor({state:'hidden'})
  await page.locator('tr[data-record-id="1"].record-highlight').waitFor()
  await page.waitForFunction(() => window.clipTest.reveals.at(-1) === false)
  assert(errors.length === 0, `运行时错误：${errors.join(';')}`)
  return {clipboardTable:true, layouts:6, trayEvents:true, errors}
}
