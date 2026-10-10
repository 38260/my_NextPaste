# 传笺 ChuanJian 桌面端功能迭代 TODO

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

- [x] T6.1 前端 `AppInner.vue`：监听窗口宽度，侧栏宽度按断点自适应（≥1200=420 / 1000-1200=380 / <1000=320）
- [x] T6.2 `LogViewerPanel.vue`：compact 模式（<1100 触发）列宽自适应，内容列放开为弹性宽度
- [x] T6.3 验证：vue-tsc + vite build 通过

## 13. 需求九：品牌整理（传笺 logo 入库 + 应用图标）（commit: chore(brand): 传笺 logo 入库并接入 README 与桌面端图标）

- [x] T9.1 抠底：生成图背景里烘焙了"假透明棋盘格"（浅灰 236 / 近白 255 交替）与投影，用「低饱和+高亮度候选 → 取与图像边框连通的连通分量」精确分离外部背景（图内浅色信笺/缎带不被掏空），再羽化 1.1px 并对白底反预乘，消除边缘浅色描边
- [x] T9.2 品牌资源入库 `docs/assets/`：logo.png（1024 主文件，RGBA 透明）/ logo-512.png（295KB）/ logo-256.png（README 用，87KB）
- [x] T9.3 应用图标：`server/build/appicon.png`（1024，四周留 6% 呼吸位）+ `server/build/windows/icon.ico`（16/24/32/48/64/128/256 七档）；`.gitignore` 增加例外把图标源文件纳入版本管理（`server/build/bin/` 仍忽略）
- [x] T9.4 README：居中头图 + 传笺命名与 Slogan + fork 归属说明；项目结构树补 `docs/assets/`；修正 Go 版本号（1.23→1.24）与失效的许可证链接
- [x] T9.5 验证：透明底在浅底/深底以及 16–256px 小尺寸下渲染自查通过（无棋盘底、无白圈、小尺寸可辨识）
- [x] T9.6 重建 exe 使图标生效：运行中的实例（用户从桌面启动）锁住 `build/bin/NextPaste Server.exe`，首次构建在"复制产物"一步失败；不杀进程，改为把运行中的 exe 与旧备份挪到 `.brand-tmp/stale/`（Windows 允许重命名/移动运行中的 exe），重建成功（12.5MB）
- [x] T9.7 图标校验：新 exe 内嵌 7 档图标与 `build/windows/icon.ico` 的 PNG 数据逐字节一致（7/7 命中），旧 exe 0/7；提取 exe 图标目视确认已换为传笺 logo

## 14. 本轮（2026-10-10 下午）需求与拆解

> 用户三项诉求 + 追加一项：① 剪切板内容独立第三列、内容默认隐藏、文本框与图片可自定义宽高；② 修复"每一条内容都出现两边"；③ 工作模式内容移到顶部、主体仍是应用日志；④ 图片点击放大预览。
> 每项小功能完成即提交，本文档同步更新。

- [x] T10.1 修复重复：`clip_ledger.go` 内容签名台账（type+md5），窗口 3s
  - 一层一条的多记（收到 / 已接收并写入 / 检测到 / 广播）→ `onLogFull` 处按签名去重，只留最先到达那条
  - 远程内容写回系统剪贴板触发本地监听回声 → `markWritten` + `isSelfEcho` 在 `onClipboardChange(Client)` 入口吞掉（不记日志也不二次广播，断掉两端来回）
  - 顺带修 `LogEntry.Timestamp` 恒为 0（原来所有行显示的都是"当前时刻"，第三列需要每条真实时间）
  - 验证：go build/vet/test 通过，5 项台账单测 PASS
- [ ] T10.2 工作模式区移至顶部条（可折叠），主体回归应用日志 — commit: feat: 工作模式区移至顶部
- [ ] T10.3 第三列「剪切板内容」面板：条目默认隐藏内容（文本打码、图片遮罩），单条眼睛按钮揭示 + 全局显示开关；日志表移除「剪贴板内容」列 — commit: feat: 剪切板内容独立第三列（默认隐藏）
- [ ] T10.4 条目尺寸自定义：文本框宽/高、图片宽/高四项可调（滑块 + 数字输入），第三列宽度可拖，配置持久化 `np:clipLayout`，一键恢复默认 — commit: feat: 剪切板条目尺寸可调
- [ ] T10.5 图片点击放大预览：modal 大图 + 缩放（适配窗口 / 1:1 / 滚轮）+ Esc 关闭 + 原始尺寸与体积标注 — commit: feat: 剪切板图片点击放大预览
- [ ] T10.6 整体测试：全量校验门 + wails build 产物 + 真机验收清单更新

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
| 2026-10-10 | T6(响应式) | 响应式布局：窗口宽度驱动侧栏（420/380/320 三档断点）与日志列宽（compact<1100 内容列弹性化），最小窗口 900px 下仍可读 | frontend/{AppInner.vue、LogViewerPanel.vue} | vue-tsc + vite build 通过；拖动窗口实测归入验收清单 |
| 2026-10-10 | T9.x | 品牌整理：传笺 logo 去棋盘底入库（docs/assets/），README 换头图并统一命名为传笺 ChuanJian；桌面端应用图标（appicon.png / windows/icon.ico）替换为传笺 logo；.gitignore 放行图标源文件 | docs/assets/*.png、README.md、server/build/appicon.png、server/build/windows/icon.ico、.gitignore | 透明底与多尺寸渲染自查通过；wails build 重建 exe，图标生效 |
| 2026-10-10 | T10.1 | 「每条内容出现两边」根因定位并修复：一条内容在单链路被多层各记一行（server `handleBinaryText` 记"收到…" + App `onClipboardReceivedBinary` 记"已接收并写入…"；本地则是"检测到…" + `BroadcastClipboardBinary` 记"广播…"），且远程内容写回系统剪贴板必然触发本地监听，monitor 的按类型 hash 守卫挡不住跨类型触发，于是同一内容在两端来回各冒一次。改为 App 层内容签名台账（`clip_ledger.go`，3s 窗口 / 64 容量）：`onLogFull` 去重只留首条，`onClipboardChange(Client)` 入口吞掉自写回声（不记不播）。顺带修 `Timestamp` 恒为 0 导致所有行显示当前时刻 | server/clip_ledger.go、clip_ledger_test.go、app.go | go build/vet 通过；go test 5 项台账单测全 PASS（去重/窗口过期重记/类型隔离/回声识别与超时/容量淘汰） |

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
| 11 | 打开日志面板查看分类筛选 | 默认只显示"剪贴板"分类日志；切"全部分类"后连接/系统日志出现；"分类"列 tag 颜色区分 |
| 12 | 手机复制一张图片 | 日志"剪贴板内容"列渲染出图片缩略图（非文字描述）；点该行复制图标 → 提示"已复制到剪贴板"，粘贴验证图片可用 |
| 13 | 点一条文本日志的复制图标 | 该文本写回系统剪贴板，并作为本地复制同步到对端 |
| 14 | 拖动窗口边缘改变宽度（1200 / 1100 / 1000 / 900 附近） | 侧栏宽度随断点变化（420/380/320），日志列宽自适应，内容不被挤压遮挡 |
