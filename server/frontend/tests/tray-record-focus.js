// 预览服务 127.0.0.1:4175；用独立 playwright-cli 会话执行。
async (page) => {
  const assert = (condition, message) => { if (!condition) throw new Error(message) }
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', entry => { if (entry.type() === 'error') errors.push(entry.text()) })
  // 浏览器自动请求的 favicon 不属于桌面应用资源。
  await page.route('**/favicon.ico', route => route.fulfill({ status: 204, body: '' }))
  await page.addInitScript(() => {
    localStorage.clear()
    window.focusTest = { callbacks: {}, scrolls: [], reads: [], reveals: [] }
    const original = Element.prototype.scrollIntoView
    Element.prototype.scrollIntoView = function (...args) {
      window.focusTest.scrolls.push(this.getAttribute('data-record-id'))
      return original.apply(this, args)
    }
    window.runtime = {
      EventsOnMultiple: (name, callback) => { window.focusTest.callbacks[name] = callback; return () => {} },
      EventsOff: () => {}, WindowMinimise: () => {}, Environment: async () => ({ platform: 'windows' })
    }
    window.go = { main: { App: {
      GetSettings: async () => ({ maxLogs: 500 }), ApplySettings: async () => {},
      GetLogs: async () => Array.from({ length: 30 }, (_, index) => ({
        level: 'INFO', category: 'clipboard', contentId: 30 - index,
        preview: '测试内容', message: '测试事件', timestamp: 1800000000000 + index
      })),
      GetServerStatus: async () => ({ isRunning: false, clientCount: 0 }),
      GetClientStatus: async () => ({ isConnected: false, serverUrl: '' }),
      GetServerClients: async () => [], GetLocalIPs: async () => ['127.0.0.1'],
      GetAutostartStatus: async () => ({ enabled: false, mode: 'server', error: '' }),
      GetLogStorePath: async () => 'isolated-test/logs.json',
      SetContentRevealed: async value => { window.focusTest.reveals.push(value) },
      GetLogContent: async id => { window.focusTest.reads.push(id); return { found: true, type: 'text', text: '测试内容' } },
      CopyLogContent: async () => {}, ClearLogs: async () => {}, HideWindow: async () => {}, Quit: async () => {}
    } } }
  })
  await page.goto('http://127.0.0.1:4175/')
  await page.locator('tr[data-record-id="1"]').waitFor()
  await page.evaluate(() => window.focusTest.callbacks['record:focus'](1))
  await page.waitForFunction(() => window.focusTest.scrolls.length === 1)
  assert(await page.locator('tr[data-record-id="1"].record-highlight').count() === 1, '首次定位未高亮')
  await page.waitForFunction(() => !document.querySelector('tr.record-highlight'))
  await page.evaluate(() => window.focusTest.callbacks['record:focus'](1))
  await page.waitForFunction(() => window.focusTest.scrolls.length === 2)
  assert(await page.locator('tr[data-record-id="1"].record-highlight').count() === 1, '相同 ID 重复点击未高亮')
  await page.evaluate(() => window.focusTest.callbacks['record:focus'](1))
  await page.waitForFunction(() => window.focusTest.scrolls.length === 3)
  assert(await page.locator('tr[data-record-id="1"].record-highlight').count() === 1, '高亮中再次点击失效')
  await page.evaluate(() => window.focusTest.callbacks['record:focus'](2))
  await page.waitForFunction(() => window.focusTest.scrolls.length === 4)
  assert(await page.locator('tr[data-record-id="2"].record-highlight').count() === 1, '切换记录未定位')
  await page.locator('.log-panel .n-select').first().click()
  await page.getByText('系统', { exact: true }).click()
  assert(await page.locator('tr[data-record-id]').count() === 0, '筛选未隐藏剪贴板记录')
  await page.evaluate(() => window.focusTest.callbacks['record:focus'](2))
  await page.waitForFunction(() => window.focusTest.scrolls.length === 5)
  assert(await page.locator('tr[data-record-id="2"].record-highlight').count() === 1, '重复定位未恢复筛选')
  assert(await page.evaluate(() => window.focusTest.reads.length) === 0, '定位不应读取隐藏内容')
  assert(errors.length === 0, `运行时错误：${errors.join(';')}`)
  return { repeatedFocus: true, highlightRestart: true, filterReset: true, hiddenContentUnread: true, errors }
}
