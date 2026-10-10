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

- [x] T6.1 全量验证：go build / vet / test（CGO_ENABLED=0）+ darwin/linux 交叉编译 + npm run build（vue-tsc）全部通过
- [x] T6.2 运行时手动回归：24 项源码接线断言探针 ALL PASS（探针已删）；GUI 真机验收清单见第 9 节
- [x] T6.3 README 功能特性补充（五项新能力）

## 10. 需求七：日志分类（commit: feat: 日志分类与默认剪贴板过滤）

- [x] T7.1 `websocket` 包：LogCallback 重构为 LogEvent 结构（Level/Message/Preview/Category/DataType/Content），既有调用零改动（log variadic → Category=connection，logClip → clipboard）；BuildPreview 导出供 app 层共用
- [x] T7.2 `app.go`：onLogFull 适配 LogEvent，内部 onLog=system，clipboard 回调走 onLogClip；LogEntry 增加 Category
- [x] T7.3 前端 types/models 增加 category；`LogViewerPanel.vue` 新增"分类"列与分类筛选（默认"剪贴板"）
- [x] T7.4 验证：go build/vet/test + vue-tsc + vite build 全部通过

## 11. 需求八：剪贴板图片渲染与一键复制（commit: feat: 日志图片缩略图与一键复制）

- [x] T8.1 `app.go`：内容暂存环（上限 20 条按序淘汰，图片 >20MB 跳过暂存），LogEntry 增加 ContentID；新增 GetLogContent / CopyLogContent 绑定（CopyLogContent 复用 clipboardMon.SetClipboard 写回系统剪贴板）
- [x] T8.2 前端 wailsjs 绑定补充；`LogViewerPanel.vue` 图片行懒加载缩略图（base64 data URL，ref 缓存防重复拉取）+ 复制图标按钮
- [x] T8.3 验证：go build/vet/test、vue-tsc + vite build 全部通过；缩略图渲染/一键复制实测归入验收清单

## 12. 需求六：桌面端响应式布局（commit: feat: 桌面端响应式布局）

- [ ] T6.1 前端 `AppInner.vue`：监听窗口宽度，侧栏宽度按断点自适应（420→360→300）
- [ ] T6.2 `LogViewerPanel.vue`：compact 模式列宽自适应，窄窗保证可读性
- [ ] T6.3 验证：vue-tsc + vite build

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
| 2026-10-10 | T6.x | 整体测试：全量校验门（build/vet/test/交叉编译/前端构建）+ 24 项接线断言探针；README 补充五项新能力 | README.md、todo.md | 校验门全过、探针 24/24 PASS（探针用后即删） |
| 2026-10-10 | T7.x | 日志分类：LogCallback 重构为 LogEvent 结构体（顺带打通 DataType/Content 携带，为需求八铺路）；分类判定放在事件源头（ws 默认 connection / logClip=clipboard / app 内部=system），不猜消息文本 | internal/websocket/{server.go、client.go、preview.go、preview_test.go}、app.go、frontend/{types、models.ts、LogViewerPanel.vue} | go build/vet/test、vue-tsc + vite build 全部通过 |
| 2026-10-10 | T8.x | 剪贴板内容暂存（20 条环，防内存膨胀）；图片日志懒加载缩略图（contentId 缓存）；一键复制走后端 SetClipboard 写回系统剪贴板（复制后等同本地复制，会正常同步到对端） | app.go、frontend/{wailsjs/*、types、LogViewerPanel.vue} | go build/vet/test、vue-tsc + vite build 全部通过；运行时实测归入验收清单 |

## 9. 验收清单（真机确认，wails build 后执行）

沙箱内无法拉起 GUI 与写注册表（reg.exe 被安全策略拦截），以下需要真机验收：

| # | 操作 | 预期 |
|---|---|---|
| 1 | `wails build` 后启动正式版，设置开启"开机启动" | reg query `HKCU\...\Run /v NextPaste` 可见 `"exe路径" -minimized -autostart=server`；重启电脑后应用静默最小化启动 |
| 2 | dev 模式（wails dev）下开"开机启动" | 报错提示"请在正式构建后开启"，注册表不写入 |
| 3 | 关闭"开机启动" | Run 键中 NextPaste 项被删除 |
| 4 | 客户端模式点击模板 tag | 地址填入并自动发起连接；已连接时提示先断开 |
| 5 | 「存为模板」→ 改名 → 保存；删除用户模板 | 用户模板可存可删，内置两个模板无删除入口，重启后仍在 |
| 6 | 手机复制一段文本 | 日志表出现"剪贴板内容"列显示文本前 120 字符（换行显示 ⏎）；手机复制图片则显示"图片 · x.xx MB" |
| 7 | 默认设置下点头部"退出" | 窗口最小化驻留任务栏、同步日志继续滚动、任务栏图标可点回；Alt+F4 同样被拦截 |
| 8 | 设置面板点"退出应用" | 进程真正退出 |
| 9 | 关闭行为切到"退出应用"后点头部按钮 | 直接退出 |
| 10 | pin 一条 Tailscale 地址，开"仅显示固定地址" | 固定地址置顶常显、虚拟网卡地址消失；重启后 pin 状态保留 |
