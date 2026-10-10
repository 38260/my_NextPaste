# NextPaste 桌面端功能迭代 TODO

> 规划文档：docs/feature-plan-2026-10-10.md · 每完成一项勾选并同步 commit

## 0. 前置

- [x] T0.1 两张流程图入库（docs/diagrams/architecture.svg、sync-workflow.svg）

## 1. 需求一：Windows 开机自启动（commit: feat: Windows 开机自启动）

- [x] T1.1 新增 `server/autostart_windows.go`：注册表 Run 键实现 setAutostart / isAutostartEnabled，写 `"exe" -minimized -autostart=<mode>`，dev 构建检测报错
- [x] T1.2 新增 `server/autostart_other.go`：非 Windows 平台沿用 go-autostart，接口同名
- [x] T1.3 重构 `server/settings.go`：删除 canUseAutostart no-op，统一模板入口
- [x] T1.4 `server/app.go` ApplySettings：AutoStartEnabled 开关注册/删键，mode 变更重注册；新增 GetAutostartStatus 供前端回读真实状态
- [x] T1.5 前端 `AppInner.vue`：watch settings 调用后端 ApplySettings（debounce 300ms），打通缺失链路
- [x] T1.6 前端 `SettingsModal.vue`：恢复"开机启动"checkbox（含使用提示文案）
- [x] T1.7 验证：go build/vet（CGO_ENABLED=0）、darwin/linux 交叉编译、TestAutostartDevGuard（dev 防护拦截+只读查询）全部通过；reg.exe 被沙箱黑名单拦截，注册表写/删真机实测归入验收清单 T6.2

## 2. 需求二：连接配置模板（commit: feat: 客户端连接配置模板）

- [x] T2.1 前端新增模板工具模块：localStorage `np:connTemplates`，首启写入内置默认模板（builtin 不可删）
- [x] T2.2 `ClientConfigForm.vue`：模板 tag 列表渲染（tooltip 展示完整地址），点击=填入 URL+未连接时自动连接，已连接时提示先断开
- [x] T2.3 `ClientConfigForm.vue`：「存为模板」内联小表单（默认名=URL host，可改名）与用户模板删除（内置不可删）
- [x] T2.4 验证：vue-tsc 类型检查 + vite build 通过；模板点击/存删/持久化运行时实测归入验收清单 T6.2

## 3. 需求三：日志剪贴板内容预览列（commit: feat: 日志展示剪贴板内容预览）

- [x] T3.1 `websocket` 包：LogCallback 扩展为 (level, message, preview)；server 广播/收文本/收图片、client 发送/收文本/收图片六处携带预览，其余日志走 variadic 兼容不动
- [x] T3.2 `app.go`：LogEntry 增加 Preview 字段，onLogFull 适配（onLog 保留为无预览包装）
- [x] T3.3 前端 `types.ts` LogEntry、`wailsjs/models.ts` LogEntry 增加 preview 字段
- [x] T3.4 `LogViewerPanel.vue`：新增"剪贴板内容"列（NEllipsis + tooltip，等宽字体，空预览显示 — 占位）
- [x] T3.5 验证：preview_test.go 单测（换行替换/rune 截断/图片大小/未知类型）全过；go build/vet、vue-tsc + vite build 通过；运行时实测归入验收清单

## 4. 需求四：关闭窗口后台常驻（commit: feat: 关闭窗口后台常驻）

- [x] T4.1 `main.go` 增加 OnBeforeClose 钩子 + `app.go` beforeClose：closeAction=minimize（默认）→ WindowMinimise 并阻止关闭；close → 放行退出
- [x] T4.2 前端 `SettingsModal.vue`：恢复"关闭按钮"设置（默认"最小化到任务栏"）+ 说明文案；设置面板底部加"退出应用"按钮（Quit 不经 OnBeforeClose，真退出保底入口）
- [x] T4.3 前端 `AppInner.vue`：头部关闭按钮按 closeAction 分流（minimize→WindowMinimise+提示 / close→Quit），tooltip 文案动态切换
- [x] T4.4 验证：go build/vet/test 全过、vue-tsc + vite build 通过；关窗/Alt+F4/切模式运行时实测归入验收清单

## 5. 需求五：连接地址固定展示 pin（commit: feat: 连接地址固定展示）

- [x] T5.1 `ConnectionInfoCard.vue`：地址行加图钉操作，localStorage `np:pinnedAddresses` 持久化
- [x] T5.2 固定地址按 pin 顺序置顶；当前枚举不到时灰态保留并标注"当前网络不在此网段"
- [x] T5.3 卡片头部"仅显示固定地址"checkbox（默认关），开启后只显示固定项（屏蔽虚拟网卡噪音）
- [x] T5.4 验证：vue-tsc + vite build 通过；pin/重启/仅显示固定的运行时实测归入验收清单

## 6. 收尾

- [ ] T6.1 全量验证：go vet / go build / npm run build
- [ ] T6.2 运行时手动回归清单（见规划文档第四节）
- [ ] T6.3 README 功能特性补充（五项新能力）

## 7. 本轮不做（含原因）

| 项 | 原因 | 前提何时失效 |
|---|---|---|
| 系统托盘图标（notification area） | Windows 托盘库（systray 系）依赖 CGO，与仓库 CGO_ENABLED=0 构建约束冲突；需求 4 的任务栏驻留方案已满足"常驻"诉求 | 若未来放开 CGO 或迁移 Wails v3（原生托盘），可重做 |
| macOS/Linux 自启实测 | 本机无对应环境，仅保证编译通过（go-autostart 非 Windows 路径无 CGO 问题） | 有 mac/Linux 环境时补测 |
| 局域网 IP 恒定（如 192.168.1.5 不变） | IP 稳定性属网络层：DHCP/MAC 绑定在路由器侧操作，或使用 Tailscale 地址（设备终身不变）；应用层无法承诺 | 不适用——需求 5 的 pin 只解决展示稳定性 |
| 服务器模式 address/port 模板化 | 需求 2 按用户确认收敛为客户端连接模板；服务器监听配置本就持久化在 localStorage，无多套切换诉求 | 用户提出"多套监听配置切换"时再扩展 |
| 协议层改动（V1.1 帧格式/V2） | 手机端是原作者上架应用商店的版本，协议改动会导致商店版客户端失联（fork 场景红线） | 自己上架鸿蒙客户端后才允许演进协议 |

## 8. 改动记录

| 日期 | 批次 | 内容与原因 | 涉及文件 | 验证 |
|---|---|---|---|---|
| 2026-10-10 | T0.1 | 两张流程图（架构拓扑 + 同步流程）与规划文档、todo 入库 | docs/diagrams/*.svg、docs/feature-plan-2026-10-10.md、todo.md | 浏览器打开 SVG 渲染正常 |
| 2026-10-10 | T1.x | 开机自启改为纯 Go 注册表实现（弃用需 CGO 的 go-autostart Windows 路径）；补上前端从未调用 ApplySettings 的断链；dev 构建防护防止把临时 exe 写进 Run 键 | server/autostart_windows.go、autostart_other.go、autostart_windows_test.go、settings.go、app.go、go.mod、frontend/{AppInner.vue、SettingsModal.vue、wailsjs/*} | go build/vet CGO_ENABLED=0 通过；darwin/linux 交叉编译通过；TestAutostartDevGuard PASS；npm run build（vue-tsc）通过 |
| 2026-10-10 | T2.x | 连接模板：内置默认（本机测试/局域网服务器）固定保留不可删，用户模板可存可删；点击模板即填入地址并自动连接；存模板走内联小表单避免弹窗打断 | frontend/src/utils/connTemplates.ts、frontend/src/components/ClientConfigForm.vue | vue-tsc + vite build 通过；运行时交互归入验收清单 |
| 2026-10-10 | T3.x | 日志管道携带剪贴板内容预览：文本按 rune 截 120 字符（换行→⏎，UTF-8 安全）、图片显示大小；websocket 内部 log 用 variadic 参数，全部既有调用零改动 | internal/websocket/{preview.go、preview_test.go、server.go、client.go}、app.go、frontend/{types、models.ts、LogViewerPanel.vue} | go build/vet、3 项 preview 单测、vue-tsc + vite build 全部通过 |
| 2026-10-10 | T4.x | 关闭窗口后台常驻：OnBeforeClose 按 closeAction 分流（默认 minimize→任务栏驻留）；真正退出保底入口放设置面板（Quit 不经过 OnBeforeClose），避免 minimize 模式下无路可退 | main.go、app.go、frontend/{SettingsModal.vue、AppInner.vue} | go build/vet/test 全过；vue-tsc + vite build 通过；关窗实测归入验收清单 |
| 2026-10-10 | T5.x | 连接地址固定：图钉 pin 持久化（np:pinnedAddresses），固定项按 pin 顺序置顶、枚举不到仍灰态保留（列表视角稳定，手机端一次输入永不用改）；"仅显示固定"一键屏蔽 VMware 等虚拟网卡噪音 | frontend/src/components/ConnectionInfoCard.vue | vue-tsc + vite build 通过；运行时实测归入验收清单 |
