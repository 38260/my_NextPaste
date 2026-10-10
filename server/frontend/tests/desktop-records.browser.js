// 先启动前端到 5179，再用 playwright-cli run-code --filename 执行；仅模拟 Wails 外部接口。
async (page) => {
  const failures = []
  page.on('pageerror', (error) => failures.push(error.message))
  await page.addInitScript(() => {
    localStorage.setItem('np:colorScheme', new URL(location.href).searchParams.get('theme') || 'light')
    const events = {}
    const logs = Array.from({ length: 30 }, (_, i) => ({
      level: 'INFO', category: 'clipboard', contentId: i + 1,
      message: i === 0 ? '收到文本数据 · 设备 https://device.example/' + 'LongDeviceIdentifier'.repeat(10) : '同步事件说明，悬浮后表格和侧栏尺寸应保持不变。'.repeat(5),
      preview: i === 0 ? 'https://example.com/' + 'LongUnbrokenClipboardText'.repeat(8) : '鼠标预览测试，悬浮后表格和侧栏尺寸应保持不变。'.repeat(5),
      timestamp: 1791633600000 - i * 1000
    }))
    window.__mock = { logs, settings: { maxLogs: 700 }, applied: [], reads: [], failClear: false }
    window.runtime = new Proxy({
      EventsOnMultiple: (name, cb) => { events[name] = cb; return () => {} },
      EventsOff: () => {}
    }, { get: (target, key) => target[key] || (() => {}) })
    window.__emit = (name, payload) => events[name]?.(payload)
    window.go = { main: { App: {
      GetSettings: async () => window.__mock.settings,
      ApplySettings: async (raw) => { window.__mock.settings = JSON.parse(raw); window.__mock.applied.push(JSON.parse(raw)) },
      GetLogs: async () => window.__mock.logs,
      ClearLogs: async () => {
        if (window.__mock.failClear) throw new Error('测试写入失败')
        window.__mock.logs = []
        window.__emit('logs:updated', [])
      },
      GetLogContent: async (id) => { window.__mock.reads.push(id); return { found: true, type: 'text', text: '完整剪贴板内容 ' + id } },
      GetLogStorePath: async () => 'D:/AIWorkSpace/WbAI/NextPaste/my_NextPaste/data/logs.json',
      GetServerStatus: async () => ({ isRunning: true, clientCount: 2 }),
      GetServerClients: async () => [{ id: 'one', deviceName: '测试手机', platform: 'HarmonyOS', connTime: '20:00' }],
      GetClientStatus: async () => ({ isConnected: false }),
      GetAutostartStatus: async () => ({ enabled: false, mode: 'server', error: '' }),
      GetLocalIPs: async () => ['192.168.1.2'], CopyLogContent: async () => {}
    } } }
  })
  const measure = () => ({
    rects: [...document.querySelectorAll('.body-log, .body-clip, .n-data-table-table, .n-data-table-tr')].map((el) => {
      const rect = el.getBoundingClientRect()
      return [rect.x, rect.y, rect.width, rect.height]
    }),
    width: document.documentElement.scrollWidth,
    height: document.documentElement.scrollHeight
  })
  const hoverResults = []
  await page.goto('http://127.0.0.1:5179/')
  await page.evaluate(() => localStorage.removeItem('np:settings'))
  for (const theme of ['light', 'dark']) {
    await page.goto('http://127.0.0.1:5179/?theme=' + theme)
    await page.waitForSelector('.n-data-table-td .n-ellipsis')
    await page.waitForFunction(() => window.__mock.applied.length > 0)
    if ((await page.locator('.n-data-table').innerText()).includes('LongUnbrokenClipboardText')) {
      throw new Error('日志表重复展示了剪贴板正文')
    }
    if (await page.evaluate(() => window.__mock.settings.maxLogs !== 700 || window.__mock.reads.length !== 0)) {
      throw new Error('未恢复后端数量设置，或隐藏内容时请求了全文')
    }
    for (const width of [900, 1200, 1500]) {
      await page.setViewportSize({ width, height: 800 })
      await page.mouse.move(0, 0)
      await page.waitForTimeout(150)
      for (const index of [0, 1]) {
        const before = await page.evaluate(measure)
        await page.locator('.n-data-table-td .n-ellipsis').nth(index).hover()
        await page.locator('.n-popover').waitFor({ state: 'visible' })
        await page.waitForTimeout(150)
        const after = await page.evaluate(measure)
        const unchanged = before.rects.length === after.rects.length && before.rects.every((rect, i) => rect.every((value, j) => Math.abs(value - after.rects[i][j]) < 0.5))
        if (!unchanged || after.width > width || after.height > before.height) {
          throw new Error(`悬浮导致布局或页面溢出: ${theme}, ${width}px, 文本${index}: ${JSON.stringify({ before, after })}`)
        }
        hoverResults.push({ theme, width, text: index, passed: true })
        if (width === 1200 && index === 0) await page.screenshot({ path: 'output/playwright/records-hover-' + theme + '.png' })
        await page.mouse.move(0, 0)
        await page.locator('.n-popover').waitFor({ state: 'hidden' })
      }
    }
  }
  await page.getByTitle('显示内容', { exact: true }).first().click()
  await page.waitForSelector('.clip-text')
  if (await page.locator('.clip-text').first().innerText() !== '完整剪贴板内容 1') throw new Error('全文未展示')
  await page.evaluate(() => { window.__mock.failClear = true })
  await page.getByRole('button', { name: '清空', exact: true }).click()
  await page.getByText('清空记录失败：Error: 测试写入失败', { exact: true }).waitFor()
  if (await page.locator('.clip-item').count() !== 30) throw new Error('清空失败时丢失了界面记录')
  await page.evaluate(() => { window.__mock.failClear = false })
  await page.getByRole('button', { name: '清空', exact: true }).click()
  await page.getByText('暂无剪切板内容', { exact: true }).waitFor()
  await page.evaluate(() => window.__emit('logs:updated', [{ level: 'INFO', category: 'clipboard', contentId: 1, message: '新记录', preview: '新预览', timestamp: Date.now() }]))
  await page.getByTitle('显示内容', { exact: true }).first().click()
  await page.waitForFunction(() => window.__mock.reads.length === 2)
  await page.evaluate(() => {
    const settings = JSON.parse(localStorage.getItem('np:settings'))
    localStorage.setItem('np:settings', JSON.stringify({ ...settings, maxLogs: 900 }))
  })
  await page.reload()
  await page.waitForFunction(() => window.__mock.applied.length > 0)
  if (await page.evaluate(() => window.__mock.settings.maxLogs !== 900)) throw new Error('用户已保存的数量设置被覆盖')
  if (failures.length) throw new Error('浏览器运行错误: ' + failures.join('; '))
  return { hoverResults, contentShownOnce: true, settingsRestore: true, fullText: true, clearFailure: true, cacheCleanup: true, savedSettings: true, pageErrors: failures }
}
