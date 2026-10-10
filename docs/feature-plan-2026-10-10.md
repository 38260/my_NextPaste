# 桌面端功能规划：开机自启 / 连接模板 / 日志内容预览 / 关闭后台常驻 / 固定连接地址

> 关闭窗口的任务栏驻留方案已由 `todo.md` 第 16 节的系统托盘方案取代；下文保留原规划记录。

> 日期：2026-10-10 · 状态：待审批 · 范围：`server/`（Wails 桌面端，Go + Vue3）

## 一、现状摸底结论

| 事项 | 现状 |
|---|---|
| 开机自启 | 设置项 `autoStartEnabled` 已存在于数据模型，但 Windows 下为 no-op：`settings.go` 依赖 `emersion/go-autostart`，其 Windows 实现（注册表）需 CGO，而仓库默认 `CGO_ENABLED=0` 无法编译；`app.go` 的 `ApplySettings` 刻意留空；前端 SettingsModal 中"开机启动"checkbox 被注释隐藏 |
| ApplySettings 调用链 | **前端从未调用过后端 ApplySettings**，设置只写 localStorage（`np:settings`）。即使后端实现了自启，开关也不会生效——此链路必须补 |
| 连接配置 | 客户端模式只有一个 `clientUrl` 输入框（存 localStorage），无多套配置管理 |
| 日志展示 | `LogEntry{level, message, timestamp}`，内容列只显示"广播剪贴板数据: text"这类描述，真实剪贴板内容在日志管道中未携带 |
| 日志管道 | `websocket.LogCallback func(level, message string)`，广播/收发日志在 `internal/websocket` 层内部打出，真实内容在该层可见但未传出 |

## 二、需求与方案

### 需求 1：Windows 开机自启动

**方案：弃用 go-autostart 的 Windows 路径，改用纯 Go 注册表实现（无 CGO）**

- 新增 `server/autostart_windows.go`（`//go:build windows`）：
  - 写 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`，值名 `NextPaste`
  - 数据：`"exe绝对路径" -minimized -autostart=server|client`（与 `flags.go` 现有参数对齐，exe 路径带引号防空格）
  - 提供 `setAutostart(enabled bool, mode AutoStartMode) error` 与 `isAutostartEnabled() (bool, error)`
  - 依赖 `golang.org/x/sys/windows/registry`（x/sys v0.33.0 已在依赖树，仅转为 direct）
- 新增 `server/autostart_other.go`（`//go:build !windows`）：macOS/Linux 继续走 `emersion/go-autostart`（非 Windows 平台无 CGO 问题），接口同名
- `settings.go`：删除 `canUseAutostart` no-op 及相关注释
- `app.go` 的 `ApplySettings`：`AutoStartEnabled=true` → 注册（mode 变更时先删后写，幂等）；`false` → 删键。返回错误给前端展示
- 前端 `SettingsModal.vue`：恢复被隐藏的"开机启动"checkbox
- 前端 `AppInner.vue`：watch settings 变化时调用后端 `ApplySettings`（补上缺失的调用链，debounce 300ms 防连点）

**风险与对策**
- `wails dev` 下 `os.Executable()` 指向临时构建产物，注册表会写错路径 → 检测到运行于 dev 构建时返回明确错误"请在正式构建（wails build）后开启"
- 用户手动清理注册表导致状态不同步 → 打开设置面板时从后端读回 `isAutostartEnabled` 刷新 checkbox 状态

### 需求 2：连接配置模板（固定默认 + 用户自定义）

**方案：纯前端实现，localStorage 持久化，不动协议与后端**

- 数据模型（`np:connTemplates`）：`[{ id, name, url, builtin: boolean }]`
- 内置默认模板（`builtin: true`，固定存在、不可删除）：
  - 「本机测试」`ws://localhost:8080/ws/my-room`
  - 「局域网服务器」`ws://192.168.1.100:8080/ws/my-room`（占位，用户可在此基础上另存自己的）
- 用户模板（`builtin: false`）：在 ClientConfigForm 中"保存当前地址为模板"，可删除
- 交互：模板列表渲染为可点击 tag；点击 = 填入 URL 输入框 +（未连接时）自动发起连接；已连接时提示"请先断开"
- 首次启动：检测 localStorage 无该 key 时写入内置默认模板

**不做**（收敛范围）：服务器模式 address/port 的模板化、模板重命名——后续有需要再扩展

### 需求 3：日志新增"剪贴板内容"列

**方案：日志管道携带内容预览，前端加列展示**

- 后端 `LogEntry` 增加 `Preview string` 字段
- `websocket.LogCallback` 签名扩展为 `func(level, message, preview string)`：
  - `server.go`：`BroadcastClipboardBinary`（"广播剪贴板数据"）、收文本/收图片完成处携带预览
  - `client.go`：`SendClipboardBinary`（"发送剪贴板数据"）、收文本/收图片完成处携带预览
  - 其余日志调用传 `""`
- `app.go`：`onLog` 适配新签名；preview 生成规则：
  - 文本：UTF-8 安全截断前 120 字符，换行替换为 `⏎`，超长追加 `…`，日志整体仍限 500 条
  - 图片：`图片 · 1.23 MB`（不落盘、不做 base64）
- 前端：
  - `types.ts` 的 `LogEntry` 增加 `preview?: string`
  - `LogViewerPanel.vue` 新增"剪贴板内容"列（约 240px，ellipsis + tooltip，等宽字体展示文本预览）

**安全考量**：预览会明文显示剪贴板内容（含可能的密码）——仅存于内存日志环（已有 500 条上限、清空按钮），与现有 message 同生命周期，不额外落盘

### 需求 4：关闭窗口后台常驻（默认驻留任务栏）

**方案：拦截"关闭"动作 → 最小化到任务栏，进程与同步持续运行；不做托盘（避开 CGO）**

- 行为定义：
  - `closeAction = "minimize"`（**默认值**，前后端现状默认即为此，无需迁移）：点击头部"退出"按钮 / 任何原生关闭路径（Alt+F4 等）→ 窗口最小化驻留任务栏，WebSocket 服务与剪贴板监听继续运行；点击任务栏图标即可恢复窗口
  - `closeAction = "close"`：行为同现在，真正退出进程
- 后端：
  - `main.go` 增加 `OnBeforeClose` 钩子（Wails v2 支持），读 `a.settings.CloseAction`：minimize → `runtime.WindowMinimise` + 返回 true（阻止关闭）；close → 返回 false 放行
  - `runtime.Quit()`（设置面板的真正退出按钮）不受 OnBeforeClose 拦截，作为 minimize 模式下的保底退出路径
- 前端：
  - `SettingsModal.vue`：恢复被隐藏的"关闭按钮"设置（radio：最小化到任务栏 / 退出应用），默认前者
  - 设置面板底部新增"退出应用"按钮（调 `Quit()`）——**关键保底**：minimize 模式下头部退出按钮已变为"最小化"，用户需要一处明确可达的真正退出入口
  - 头部"退出"按钮：minimize 模式下执行 `WindowMinimise` 并 tooltip 提示"已最小化到任务栏（后台保持同步）"；close 模式下维持 `Quit()`
- **不做托盘**：任务栏驻留方案无需托盘图标；托盘（notification area）需 CGO 依赖的 systray 库，与仓库 CGO_ENABLED=0 约束冲突，留作后续可选

**风险**：minimize 模式下用户找不到退出入口 → 已通过"设置面板退出按钮"解决；后台运行时用户可能忘记应用在跑 → 设置项旁加一句说明文案

### 需求 5：连接地址固定展示（pin）

**背景**：可用连接地址为运行时枚举本机全部网卡 IP（`GetLocalIPs`），跨网络切换/DHCP 重分配/虚拟网卡（VMware 等）都会导致列表跳动；用户希望地址稳定，手机端只输一次。

**前提认知（应用层无法保证的事）**：
- 局域网 IP（如 192.168.1.5）是否恒定取决于网络层——Tailscale 组网 IP（100.64.x.x）设备终身不变，推荐作为固定地址；或路由器做 DHCP/MAC 绑定
- 端口号本身已由 localStorage 持久化，稳定

**方案：纯前端 pin 机制，不动后端枚举逻辑**

- `ConnectionInfoCard.vue` 每条地址增加"固定/取消固定"图钉操作，固定地址持久化到 localStorage（`np:pinnedAddresses`，存完整 ws 地址）
- 固定地址**永远置顶**并带"已固定"标记；即使当前枚举不到该 IP 也照常显示（灰态标注"当前网络不在此网段"），保证列表视角稳定
- 新增"仅显示固定地址"开关（默认关）：开启后过滤全部实时枚举项（屏蔽 VMware/虚拟网卡噪音）
- 排序规则：固定地址（pin 顺序）→ 实时枚举地址（原顺序）

**验证**：pin 后重启应用列表稳定；"仅显示固定"后噪音地址消失；换网段后固定地址灰态保留。

## 三、实施顺序与提交计划

按"一个功能一次提交"：

| # | 内容 | commit |
|---|---|---|
| 0 | 已保存的两张流程图入库（`docs/diagrams/`） | `docs: 新增架构与同步流程图` |
| 1 | 需求 1：开机自启（Go 注册表实现 + 设置链路打通 + 前端开关恢复） | `feat: Windows 开机自启动` |
| 2 | 需求 2：连接配置模板 | `feat: 客户端连接配置模板` |
| 3 | 需求 3：日志剪贴板内容预览列 | `feat: 日志展示剪贴板内容预览` |
| 4 | 需求 4：关闭窗口后台常驻 | `feat: 关闭窗口后台常驻` |
| 5 | 需求 5：连接地址固定展示（pin） | `feat: 连接地址固定展示` |

依赖关系：五个功能相互独立，无技术依赖，按上表顺序串行执行；需求 1/4、需求 2/5 分别触碰同批文件（SettingsModal/AppInner、连接相关组件），相邻顺序执行减少冲突。

## 四、验证方案

- **静态**：`go vet ./...`、`go build ./...`（CGO_ENABLED=0，验证无 CGO 依赖回归）
- **前端**：`npm run build`（含 vue-tsc 类型检查）通过
- **运行时**（wails dev 手动验证清单）：
  1. 设置开"开机启动"→ `reg query HKCU\...\Run /v NextPaste` 确认写入与参数正确；关 → 键被删除
  2. dev 模式下开启自启应得到明确报错提示
  3. 客户端模式下模板点击/保存/删除（内置模板删不掉）
  4. 复制文本/图片后日志表出现"剪贴板内容"列，文本显示前 120 字符、图片显示大小
  5. 默认（minimize）下点头部"退出"→ 窗口最小化驻留任务栏、任务栏可点回、同步日志持续滚动；Alt+F4 同样被拦截；设置面板"退出应用"可真正退出；切到"退出应用"模式后头部按钮直接退出
  6. pin 连接地址后重启应用列表稳定不变；"仅显示固定地址"后虚拟网卡地址消失；Tailscale 地址跨网络可用
- **回归**：服务器模式启动/停止、客户端连接/断开、心跳日志均正常

## 五、已知限制

- macOS/Linux 的自启走 go-autostart 原路径，本次不实测（无环境），仅保证编译通过
- `-minimized` 静默启动 + 无托盘：开机自启后窗口隐藏、仅任务栏驻留——需求 4 落地后任务栏图标即为找回入口；托盘（notification area）需 CGO 依赖的 systray 库，与 CGO_ENABLED=0 约束冲突，本次不做，留作后续可选

