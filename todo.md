# 传笺 ChuanJian 桌面端功能迭代 TODO

> 规划文档：docs/feature-plan-2026-10-10.md · 每完成一项勾选并同步 commit

## 29. README 同步与推送 GitHub（2026-10-11）

- [x] T29.1 根 README 桌面端特性补充单实例、托盘右键菜单（最近记录/开机自启开关）、绿色图标分工；快速开始标注产物 `ChuanJian.exe`，启动说明对齐 T28 后现状（手动启动显示主窗口、开机自启才隐藏到托盘）。server/README 新增系统集成段（托盘常驻/单实例/开机自启/本地记录）与剪贴板内容表，构建产物与使用说明同步，Go/Node 版本要求对齐 go.mod。提交后推送 GitHub `origin/main`。提交：`docs: 更新 README 同步更名与托盘新特性`。

## 28. 桌面快捷方式与真实托盘点击修复（2026-10-11）

- [x] T28.1 根因：旧代码只拦截 GetMessage 返回的队列消息，直接送入窗口过程的点击与菜单命令绕过应用。补充 SendMessage 用例后先复现左键回调超时；改用同线程 SetWindowSubclass 统一处理两条路径，安装回调后才标记托盘就绪。直接左/右键、直接设置命令、队列左右键及菜单重复打开/取消/退出回归通过。提交：`fix: 托盘窗口过程处理真实点击和菜单命令`。
- [x] T28.2 桌面快捷方式参数清空，手动启动正常显示窗口，自启仍保留 `-minimized -autostart=server`；经设置面板正常退出旧版后留备份、安装新版并显示启动。Go 全量测试/静态检查与 Wails 生产构建通过；用户物理确认“菜单能弹出，打开设置也正常”。正式程序 SHA256：`D90A4F0C85AF24AC3D44B07B003C5784E75E90C7A361147C733539D17806ACDE`。详见 `docs/tray-click-acceptance-2026-10-11.md`。提交：`docs: 记录真实托盘点击与桌面启动验收`。

## 27. 审阅问题修复（2026-10-11）

- [x] T27.1 WebView2 迁移失败时返回错误并停止启动，通过 Win32 对话框提示关闭旧版/检查权限后重试；不初始化新目录，也不继续迁移自启。目录占用后重试、已有目标数据保留、空路径/文件冲突测试与 `go vet` 通过。提交：`fix: 迁移失败时阻止初始化新用户数据目录`。
- [x] T27.2 每次托盘记录事件直接调用表格定位方法，重建高亮并清理卸载时的计时器；同 ID 等待后重击、高亮期间重击、切换 ID、筛选隐藏后再次定位、隐藏内容不读取的浏览器回归通过，运行时错误 0；前端类型检查/构建通过。提交：`fix: 连续点击同一托盘记录时重新定位并高亮`。
- [x] T27.3 自启迁移先保证新值成功写入/已存在，再删除旧值，已有新配置保持不变；独立注册表测试键覆盖无旧项、server/client 模式、已有新值、真实写入失败/删除失败及恢复权限后重试，6 组测试与开发模式防护、`go vet` 通过。提交：`fix: 自启迁移先写后删并保留已有配置`。
- [x] T27.4 补齐 `SetContentRevealed` Mock，捕获 Vue 控制台错误并保留 pageerror 断言；验证隐藏偏好、记录定位、窗口恢复和自启开关/保存事件，更新设置帮助文字。完整表格回归（全文/图片/复制/缓存、6 组布局、筛选/隐藏、清空竞态与失败保护）通过，错误 0；前端类型检查/构建通过。提交：`test: 修复浏览器回归漏检并覆盖托盘事件`。
- [x] T27.5 Go 全量测试/静态检查、前端类型检查/构建、两项浏览器回归、Linux/macOS 交叉编译、Wails 生产构建通过。正式 exe 的原生启动失败提示/退出、单实例、托盘左键恢复/右键菜单取消、自启勾选与迁移验证通过；旧实例优雅退出后留备份并安装新版，旧自启值已删除，现有新自启配置保持不变。新版 SHA256：`F392F2B4184E8F0F728F2771E262F62A169266A51553F92C2A225EA131139E67`，PID 27456，8080 监听恢复。临时文件清理，完整证据见 `docs/fix-acceptance-2026-10-11.md`。提交：`docs: 记录审阅问题修复与新版整体验收`。

## 26. 独立审阅与验收（2026-10-11）

- [x] T26.1 核对本轮 11 个提交、运行程序 SHA256、单进程/8080、快捷方式与迁移现状；保留用户 `logo.png`。
- [x] T26.2 桌面端 Go 全量测试/静态检查、前端类型检查/构建、中继端编译与静态检查通过。
- [x] T26.3 隔离复现迁移失败和相同记录重复定位问题，完成浏览器回归、独立 Wails 生产构建与跨平台编译；验收 exe 与正式程序哈希一致。临时测试和构建产物已清理；审阅结论、失败证据和限制详见 `docs/review-acceptance-2026-10-11.md`。提交：`docs: 记录传笺桌面端独立审阅与验收结论`。
- [x] T26.4 四项问题已由 T27 修复并通过自动复验，旧自启值已清理，新自启配置保留。手机端同步、HAP 构建、物理鼠标与重启登录后的自启仍待用户验收。

## 25. 项目更名为传笺 ChuanJian（2026-10-11）

- [x] T25.1 桌面端更名与遗留迁移：wails.json 更名（name 传笺、产物 `ChuanJian.exe`、新增版本信息“传笺 ChuanJian 1.0.0”）；窗口标题/前端页面标题/头部品牌改为“传笺”；WS 客户端设备名“传笺 桌面端”；单实例 UniqueId `ChuanJian-Desktop`。新增 `brand_windows.go`：WebView2 数据目录固定到 `%AppData%\ChuanJian`（与 exe 文件名解耦），启动时自动迁入旧 `%AppData%\NextPaste Server.exe` 数据目录，exe 改名不再丢失 localStorage 配置；自启注册表值名 `NextPaste`→`ChuanJian`，启动时自动迁移旧值（保留开启状态与工作模式、路径更新为新 exe，开发构建跳过）；非 Windows 自启服务名同步更名。全量 Go 测试/静态检查、Linux/macOS 交叉编译、前端构建、Wails 生产构建通过。提交：`feat: 项目更名为传笺 ChuanJian 并迁移品牌遗留`。
- [x] T25.2 文档与鸿蒙端更名：server/README、relay-server/README 与中继服务界面横幅改为传笺（保留上游 NextPaste fork 署名、NPBP 协议名与旧记录路径迁移代码）；HarmonyOS 应用名/标题/关于页改为“传笺”（en 资源 “ChuanJian”），关于页仓库地址指向本项目 fork。提交：同 T25.1。
- [x] T25.3 交付验收：旧实例经托盘菜单优雅退出（未强制终止）后删除旧名 exe（历史备份保留）；桌面快捷方式重建为 `传笺.lnk`（指向 `ChuanJian.exe`，保留 `-minimized` 参数、工作目录与绿色图标），移除旧 `NextPaste.lnk`。启动后实测：旧注册表值已移除、新值 `ChuanJian` 指向新 exe 且自启状态/模式保留；WebView2 数据目录已迁移（8080 自动监听证明 localStorage 配置完好）；真机 Win32 验收（单实例/关闭隐藏/左键恢复/右键菜单/退出）通过；exe 版本信息经 Win32 API 核验为“传笺 ChuanJian 1.0.0”。正式程序 `server/build/bin/ChuanJian.exe` SHA256：`A9E4387F7B13243B8DA97293A40C0436ABF14B3F54543C7B3141B850899F1BE0`。临时验收脚本已清理。提交：`docs: 记录项目更名交付与迁移验收`。

用户验收：桌面“传笺”图标启动，任务栏显示“传笺”，exe 属性显示“传笺 ChuanJian 1.0.0”；连接配置与显示偏好保留、开机自启继续生效（注册表值名变为 ChuanJian）；托盘右键菜单、单实例行为不变。鸿蒙端资源更名未做真机构建验证（本机无 DevEco），下次构建 HAP 时在设备上生效。

## 24. 托盘右键记录菜单、单实例与绿色图标（2026-10-11）

- [x] T24.1 托盘左键恢复主界面并收起设置面板：左键不再触发 `settings:open`，改为发送 `window:restored`，前端收起上次遗留的设置面板，消除“左键看起来打开设置”的错觉。前端类型检查/构建通过。提交：`feat: 托盘左键恢复主界面并收起设置面板`。
- [x] T24.2 托盘右键自绘菜单：右键消息改为在托盘循环线程弹出 Win32 自绘菜单（`TrackPopupMenu`），每次弹出前按当前记录重建，展示“最近 3 条剪贴板记录 / 打开设置 / 退出应用”；记录项标题为“时间 · 摘要”（文本截断 32 rune，图片显示大小），命令 ID 从 0x7000 起与托盘库内部项隔离。点击记录发送 `record:focus`，前端关闭设置面板、必要时重置筛选、滚动定位到日志表对应行并短暂高亮，复制沿用表内既有按钮；不足 3 条按实际数量展示，清空/淘汰后弹出时自然更新。真实 Win32 队列子进程测试：25 次左键转发、左键不弹菜单、右键弹菜单并可取消、自绘命令拦截转交、非自绘命令直通托盘库、菜单可重复打开全部通过；`recentClipRecords` 数据源（最新在前、系统日志过滤、ID 引用、截断、清空后为空）单测通过。提交：`feat: 托盘右键菜单展示最近剪贴板记录并支持定位`。
- [x] T24.3 内容隐藏偏好同步：新增 `SetContentRevealed` 绑定（`wails generate module` 重新生成 wailsjs），前端把“显示内容”开关实时同步后端；关闭时托盘菜单记录标题只显示“内容已隐藏”，开启后展示摘要。子进程与单测回归通过。提交：`feat: 托盘菜单沿用内容隐藏偏好`。
- [x] T24.4 单实例：接入 Wails `SingleInstanceLock`（UniqueId `NextPaste-Server-ChuanJian`），重复启动的第二实例向首实例发送 WM_COPYDATA 后自行退出，首实例回调 `onSecondInstanceLaunch` 显示主窗口。真机验收：再次启动 exe 后仅剩 1 个进程且窗口被唤起。提交：`feat: 接入 Wails 单实例锁避免重复启动`。
- [x] T24.5 图标分工：`chuanjian-logo.png`（绿色）检测棋盘格底纹边界后加圆角透明蒙版生成 `server/build/appicon.png`（896×896 RGBA），并按 wails 规格（256/128/64/48/32/16）重生成 `build/windows/icon.ico`（wails 仅在 ico 缺失时才从 appicon 重生成，已提交新 ico）；exe 内嵌图标经 `ExtractAssociatedIcon` 核验为绿色，托盘/任务栏/桌面快捷方式随 exe 图标统一。现有蓝色品牌标识经四角泛洪去白底后存为 `docs/assets/logo-brand.png` 并替换根 README 项目介绍头图。提交：`feat: 桌面端启用绿色图标并在项目介绍展示品牌标识`、`fix: 重新生成绿色托盘与应用图标 ico`。
- [x] T24.6 整体测试与交付：全量 `go test -count=1 ./...`（含托盘子进程回归）与 `go vet ./...` 通过，前端 vue-tsc/Vite 构建通过，Windows Wails 生产构建通过。旧实例（003EF6…）通过其托盘“退出应用”菜单命令优雅退出（未强制终止），备份为 `server/build/bin/NextPaste Server.pre-menu-instance-20261011.exe`；新版安装至原路径。真机 Win32 验收：单实例重复启动仅剩 1 进程并唤起窗口、WM_CLOSE 隐藏到托盘、托盘左键恢复、右键弹出记录菜单并可取消收起、菜单“退出应用”命令优雅退出，全部通过。最终以 `-minimized` 重启，托盘驻留且 8080 监听恢复。临时验收脚本（live_acceptance_test.go、图标处理脚本）已清理，工作区仅保留用户 `logo.png`。提交：`docs: 记录托盘菜单单实例与图标交付`。
- [x] T24.7 追加：托盘右键菜单增加“开机自启”开关（用户中途补充需求）：菜单在“打开设置”上方新增勾选项，勾选态每次弹出时读注册表（以系统为准）；点击切换 HKCU Run 键，开启沿用当前工作模式，失败写入应用日志；成功后发送 `autostart:changed` 同步前端设置面板开关。全量 Go 测试/静态检查、前端构建、Wails 生产构建通过；重新安装后真机回归（单实例、关闭隐藏、左键恢复、右键菜单、退出）通过，并经真实注册表核验：切换后 Run 键正确移除/恢复且内容与原状态一致（`"exe" -minimized -autostart=server`）。最终正式程序 SHA256：`D82CB1427543558DE470298F055E46AF5CC47DC458723CDE3B3A3AB1AFE5371E`。提交：`feat: 托盘右键菜单增加开机自启开关`。

用户验收：左键托盘图标应恢复主界面且设置面板收起；右键应弹出“最近 3 条剪贴板记录 / 开机自启（勾选态）/ 打开设置 / 退出应用”，点击记录应定位到日志表对应行（隐藏偏好开启时菜单只显示“内容已隐藏”，表格内点眼睛可见后复制）；重复双击快捷方式只保留一个实例并唤起窗口；托盘切换“开机自启”后设置面板开关同步变化；exe/任务栏/托盘为绿色图标，README 头图为蓝色品牌标识。物理鼠标操作、托盘菜单实际文案与手机端联机同步待用户确认。

## 23. 托盘左键窗口、右键设置（2026-10-10）

- [x] T23.1 区分托盘左右键通知：左键直接恢复主窗口，右键恢复窗口后发送 `settings:open`，前端监听并打开已有设置面板；重复触发保持同一面板，卸载时取消监听。更新设置说明与 README，退出入口继续使用设置中的“退出应用”。全量 Go 测试/静态检查、前端类型检查/构建与浏览器回归通过；真实托盘队列 25 次左键、3 次右键及右键不弹旧菜单检查通过，浏览器设置打开/重复右键/退出入口与既有剪贴板回归通过。提交：`feat: 托盘左键恢复窗口右键直接打开设置`。
- [x] T23.2 整体测试与安装：全量 Go 测试（25 次左键/3 次右键、右键不弹旧菜单、退出及生命周期）、静态检查、前端类型检查/构建、浏览器既有表格回归和设置事件检查、Windows Wails 生产构建通过。正常退出旧实例并备份为 `server/build/bin/NextPaste Server.pre-tray-settings-20261010.exe`，新版安装至原路径并重启；安装后真实 Win32 消息验证左键恢复、关闭隐藏、右键恢复，再由 Windows 原生辅助功能树确认“系统设置”实际显示，8080 监听已恢复。正式程序与构建版 SHA256 一致：`003EF6F5994B93B67CEEB1CD584A005E70CE2B449F5B83FB341CA03D5C400345`。原有 5 个生成文件修改及 `logo.png` 保留，5 个文件内容哈希未变；测试浏览器与预览服务已关闭，本轮无新增临时脚本。提交：`docs: 记录托盘左右键设置验收与正式程序更新`。

用户验收：关闭窗口后左键托盘图标应直接恢复主窗口；右键应直接进入“系统设置”，无需再选菜单；设置中的“退出应用”应正常结束程序。自动化使用真实托盘消息和安装程序的辅助功能树核验，物理鼠标操作与手机端联机同步本轮未重测。

## 22. 剪贴板内容并入日志表（2026-10-10）

- [x] T22.1 用“剪贴板内容”替换事件列，新增单元格组件承接文本/图片、遮罩与揭示、复制、尺寸和图片预览；复用既有内容读取/复制接口与本地偏好，隐藏时不读取正文，内容删除/切换后释放单元格缓存。前端类型检查/构建通过，浏览器表内全文与按需读取检查通过。提交：`feat: 将剪贴板内容展示并入日志列`。
- [x] T22.2 移除右侧独立内容面板、拖拽宽度与关联类型/配置，日志区占满主体；更新 README。保留既有尺寸和揭示偏好，旧记录仅剩预览时也遵守全局隐藏。新增可复用浏览器断言回归，类型检查/构建、单表布局、全文/图片/复制、缓存、图片预览、尺寸持久化、900/1200/1500px × 深浅主题、清空后的延迟读取与清空失败保护均通过。提交：`refactor: 移除独立剪贴板面板并保留表内内容交互`。
- [x] T22.3 整体测试与交付：全量 Go 测试（包含真实托盘消息回归）、静态检查、vue-tsc/Vite 与 Windows Wails 生产构建通过；浏览器断言回归覆盖单表布局、隐藏时不读取、全文换行、图片放大/Esc、复制正确 ID、内容缓存、尺寸/揭示偏好持久化、900/1200/1500px × 深浅主题、分类筛选、旧预览遮罩、清空/延迟读取及失败保护，运行时异常 0。已正常退出旧实例（未强制终止），备份到 `server/build/bin/NextPaste Server.pre-content-table-20261010.exe`，新版安装至原路径并重启，托盘恢复窗口和 8080 监听正常。正式程序与构建版 SHA256 一致：`1D4F562BE2393FFD595EE8941C03C091C9936E313BECAA6325234A86484761AC`。原有 5 个生成文件修改及 `logo.png` 保留，构建前后 SHA256 未变；隔离浏览器及预览服务已关闭，保留回归脚本与本地验收截图 `output/playwright/clipboard-table-light.png`。提交：`test: 完成剪贴板表格回归与正式程序验收`。

用户验收：日志表列为“级别 / 分类 / 剪贴板内容 / 时间”，右侧没有独立剪贴板面板；点击单条眼睛或“显示内容”开关，在表内查看文本和图片，测试复制、图片放大/Esc、尺寸调节、筛选与清空。手机端真实联机同步本轮未重测；浏览器交互回归仅替代 Wails 外部接口。

## 21. 托盘点击与右键修复（2026-10-10）

- [x] T21.1 固定托盘消息循环的系统线程，避免协程迁移后窗口与消息队列分离；`go test -count=1 ./...`、`go vet ./...` 通过。提交：`fix: 固定托盘消息循环线程以恢复右键响应`。
- [x] T21.2 左键通知转交 `ShowWindow`，右键及菜单命令继续由现有 systray 库处理，沿用显示/取消最小化与退出逻辑；新增真实 Win32 窗口队列子进程测试，25 次左键（穿插 GC/调度）、3 次右键菜单打开、显示/退出菜单命令和托盘关闭均通过。提交：`fix: 接入托盘左键恢复并验证右键菜单`。
- [x] T21.3 修复正式程序回归时发现的 `-minimized` 启动竞态：Wails 导航完成后会再次显示窗口，改用 `StartHidden`，托盘不可用时先显示再最小化以保留任务栏入口。全量 Go 测试/静态检查通过；隔离配置与记录目录的正式 exe 普通/隐藏启动、各 3 轮左键恢复/关闭隐藏、右键菜单、菜单恢复与正常退出共 21 项检查通过。提交：`fix: 保持托盘隐藏启动并保留不可用时的恢复入口`。
- [x] T21.4 整体测试与安装：`go test -count=1 ./...`、`go vet ./...`、前端 vue-tsc/Vite 和 Windows Wails 生产构建通过；正式程序隔离目录 21 项 Win32 消息回归通过。旧实例通过退出菜单正常关闭（未强制终止），备份为 `server/build/bin/NextPaste Server.pre-tray-fix-20261010.exe`，新版安装到原路径 `server/build/bin/NextPaste Server.exe` 并以 `-minimized` 重启；安装后核验隐藏启动、右键菜单与左键恢复均通过，唯一新版实例恢复 8080 监听，主窗口留给用户验收。安装版与验收版 SHA256：`DC96AA3BF047476B95C096FAA32B177F28E18B315FF4DEAD3D1BC0B78B832689`。工作区原有 5 个生成文件修改与 `logo.png` 保留，5 个文件构建前后 SHA256 均未改变。临时脚本已清理；自动审批拦截递归删除，隔离测试程序/配置/记录仍保留在 Git 忽略的 `output/tray-fix-test/`，与正式程序数据独立。提交：`docs: 记录托盘修复整体验证与正式程序更新`。

用户验收：点 × 隐藏窗口后，直接左键点托盘图标应恢复；右键应出现“显示窗口 / 退出应用”，分别验证恢复与完整退出，再从原桌面快捷方式启动确认静默托盘驻留。自动测试使用真实 Win32 消息队列，物理鼠标点击及手机端同步仍待用户确认；本轮未配置 macOS/Linux GUI 环境。

## 20. 本地保存完善（2026-10-10）

- [x] T20.1 全文与图片分别落盘、日志格式兼容迁移、恢复内容 ID；增加全文/图片重启恢复、超过 20 条保留、旧格式 700 条迁移、损坏原文件保护测试；`go test -count=1 ./...` 通过。
- [x] T20.2 保存位置改到项目 `data/`（已加入 Git 忽略），从程序位置定位项目根目录、兼容复制旧位置记录且保留原文件；数量设置在启动前恢复、超量按序淘汰并清理内容文件、清空同步落盘；路径/迁移、700 条重启、缩小至 50 条、清空重启、写入失败保留原文件、无关文件保护测试通过；`go test -count=1 ./...` 通过。
- [x] T20.3 前端设置恢复与缓存清理、更新使用说明；首次读取记录文件中的数量设置，已有用户设置继续沿用；删除记录释放前端缓存，读取/清空失败显示错误；前端类型检查和构建通过，浏览器设置恢复/全文/清空失败/缓存清理回归通过。
- [x] T20.4 修复日志文字悬浮导致组件变化：限制提示框宽高、长文字自动换行；原先 1200px 页面被撑至 1426px，修复后 900/1200/1500px × 深浅主题 × 中文/连续长文本共 12 组悬浮回归通过，表格与侧栏尺寸稳定且页面无新增溢出。
- [x] T20.5 整体校验：`go test -count=1 ./...`、`go vet ./...`、前端类型检查与构建、Windows Wails 正式构建、Linux/macOS 交叉编译全部通过。对正式前端构建进行浏览器回归：悬浮布局 12/12、正文只展示一次、数量设置恢复/已有设置保留、全文按需读取、清空失败反馈与缓存释放均通过，运行错误 0。验收程序：`server/build/bin/records-acceptance/NextPaste Server.exe`，与构建产物 SHA256 一致；自动审批阻止替换已有 exe，原程序未改动，需完全退出旧版后运行此验收版。未实测原生 WebView2 与手机联机；截图仅保留本地 `output/playwright/records-hover-{light,dark}.png`。
- [x] T20.6 正文只在右侧展示，日志改为事件说明，消除日志与内容面板各出现一次的重复展示（按用户澄清，无需修改同步协议）；浏览器断言日志中不再显示剪贴板正文通过。
- [x] T20.7 完善 Git 忽略规则，本地数据/缓存/临时文件/构建产物不上传，停止跟踪验收截图但保留本地文件；8 类忽略目标检查通过，图标及构建配置保留。
- [x] T20.8 清空/缩减先落盘成功再改变内存，写入失败保留原记录和原文；新增失败回归通过，整体 Go 测试通过。
- [x] T20.9 最终验证后已推送 GitHub `origin/main`（功能与验收提交 `490a819`）；忽略项未纳入当前提交，原有未提交生成文件修改与根目录 `logo.png` 已保留在本地。

## 19. 四项需求复核（2026-10-10）

> 复核提交 `2734eea` 的实际实现。本节结论优先于第 18 节的完成标记；本轮仅检查，不修改产品代码。

- [x] T19.1 Logo 接入核对：标题栏、应用图标、托盘资源均已接入根目录 `logo.png` 的图形标识版本；裁掉名称与标语，并非原图完整照搬。
- [x] T19.2 本地记录及数量上限核对：日志元数据与预览写入 `%AppData%\NextPaste\logs.json`，默认 500 条、可设 50–10000 条，运行中超量删除最早记录。但完整剪贴板数据持久化及大于 500 条的重启恢复未完成，见下方待修项。
- [x] T19.3 服务器设备展示核对：后端统计当前连接，前端每 2 秒刷新，显示本机服务器身份、设备数量及悬浮明细。
- [x] T19.4 日志列核对：原「内容」列已替换为「剪贴板内容」，展示预览；无预览时回退事件描述。图片与全文仍由右侧剪切板面板按需加载。
- [x] T19.5 验证：现有 `go test -count=1 ./...`、`go vet ./...` 与 `npm run build` 通过（前端仍有既有大包体提示）。临时隔离探针使用测试数据和临时 APPDATA，复现以下三个问题；探针已清理，未触碰真实记录。桌面实机及手机连接未重测。
- [x] T19.6 完整内容持久化：原问题为内容仅在内存中固定保留 20 条；已由 T20.1/T20.2 修复，全文与图片分别落盘并跟随记录数量保留。
- [x] T19.7 内容 ID 恢复：原问题为重启后新内容复用旧 ID；已由 T20.1 修复，恢复 ID 计数并清除旧版缺失原文的关联。
- [x] T19.8 启动裁剪顺序：原问题为设置 700 条仍先按默认 500 条裁剪；已由 T20.2/T20.3 修复，启动前恢复数量设置。

## 18. 本轮（2026-10-10 晚）：品牌图标 / 记录本地化 / 服务器设备数 / 日志内容列

> 用户四项诉求：① 用项目根目录 `logo.png` 替换桌面端图标与标题栏 logo；② 记录放本地、可自定义数量、超量按顺序删最早；③ 展示"本机作为服务器 + 已连接设备数"；④ 日志表"内容"列替换为"剪贴板内容"列。

- [x] T18.1 品牌资源：从根目录 `logo.png`（1254×1254 白底，含"传笺 ChuanJian"名称与标语）提取**图形标识**——按最小通道做白底键出（奶白色信笺不被掏空）、反预乘消边缘白圈、紧致裁剪，得 912×499 透明标识
  - 应用图标 `server/build/appicon.png`（1024，圆角白底贴片，标识居中留 7% 呼吸位）+ `server/build/windows/icon.ico`（16/24/32/48/64/128/256 七档）
  - 标题栏标识 `server/frontend/src/assets/images/logo-mark.png`（234×128 透明，2 倍图供 28px 显示）
  - 只取图形部分：完整 logo 含中英文名称与标语，在 56px 标题栏与 16–32px 图标尺寸下文字不可辨
  - 验证：新 exe 内嵌 7 档图标与 `windows/icon.ico` 的 PNG 数据逐字节一致（7/7），旧 exe 0/7
- [x] T18.2 标题栏接入 logo：`AppInner.vue` 头部在应用名前加 28px 标识（保留 "NextPaste Server" 文案与 NPBP 标签，未改应用命名）
- [x] T18.3 记录本地持久化 + 可配置数量（超量按序删除）
  - 新增 `server/logstore.go`：路径 `%AppData%\NextPaste\logs.json`（用户配置目录，不动系统全局配置）；原子写（临时文件 + rename）；文件缺失视为空、损坏退回空并报错但不阻断启动
  - `app.go`：`startup` 加载历史记录、`shutdown` 落盘；`onLogFull` 走 800ms **节流**合并窗口异步落盘（非防抖，持续日志也能写出）；`ClearLogs` 清空后立即落盘
  - `Settings` 增加 `maxLogs`（默认 500，范围 50–10000）；`ApplySettings` 归一化并**即时裁剪**（缩小上限立刻删最早记录）
  - 前端：设置面板新增「记录与数据」区（「记录数量」数字框 + 「存储位置」路径），设置随 `np:settings` 持久化并提交后端
  - 验证：7 项单测 PASS（存取往返 / 缺失文件 / 损坏文件 / 超量裁剪 / 未超量不裁剪 / 立即落盘 / 空路径不 panic / 合并窗口内落盘）
- [x] T18.4 服务器已连接设备展示
  - 后端：`App.GetServerClients()` 暴露 `ws.Server.GetClients()`（设备名 / 平台 / 连接时间）
  - 前端：表头标签由「客户端 N」改为「已连接 N 台设备」并支持悬浮看明细；顶部条摘要改为「本机作为服务器 · 端口 X · 已连接 N 台设备」
  - 顺带修：状态刷新由"并发触发"改为"按序刷新"（设备列表依赖服务器状态），否则首屏设备明细为空
- [x] T18.5 日志表列调整：第 3 列由「内容」（事件描述）替换为「剪贴板内容」（`preview`）；无预览的记录（连接/系统）回退显示事件描述，避免整列空白；右侧剪切板面板保持不变
- [x] T18.6 整体校验：`go build` / `go vet` / `go test -count=1`（CGO_ENABLED=0）、darwin+linux 交叉编译、`npm run build`（vue-tsc + vite）、`wails build` 产出 12.8MB exe 全部通过
  - 无头 Chrome + CDP（注入 Wails 模拟接口）11/11 PASS：标题栏 logo 尺寸 28px 且资源加载成功、列头为「级别/分类/剪贴板内容/时间」、预览与回退渲染正确、表头与顶部条设备数文案正确、悬浮出设备明细、`np:settings` 含 maxLogs=500、设置面板含「记录数量/存储位置」且默认 500
  - 真机待验收项见第 9 节 24–27 项

## 17. 桌面快捷方式切换到托盘版（2026-10-10）

- [x] T13.1 更新现有桌面 `NextPaste.lnk`：使用 `-minimized` 隐藏启动，工作目录与图标同步切换，原快捷方式位置不变。发现 exe 文件名决定 Wails 用户配置目录后，最终目标改回原文件名 `server/build/bin/NextPaste Server.exe`（该文件已替换为托盘版）。提交：`docs: 记录桌面快捷方式切换`。
- [x] T13.2 通过应用设置将旧版切为真正退出并平稳关闭；旧 exe 备份为 `NextPaste Server.pre-tray.exe`，将已验收的托盘版复制到原文件名以沿用既有配置。由快捷方式启动后自动恢复服务器监听，并将关闭行为恢复为“隐藏到系统托盘”；注册表开机启动项已由应用更新到新版路径。提交：`docs: 记录托盘版实例与配置切换`。
- [x] T13.3 整体验证：回读快捷方式目标、`-minimized` 参数和图标；原文件名下的程序与托盘验收版 SHA256 一致，旧版备份保留；仅有一个新版进程，8080 端口由该进程监听，HKCU Run 项指向新版；设置中“隐藏到系统托盘”已选中，点 × 后窗口列表为空而进程与服务持续运行。同字节验收版此前已通过托盘菜单恢复 / 退出实测。提交：`docs: 完成桌面托盘常驻验收记录`。

## 16. 关闭窗口后托盘驻留（2026-10-10）

- [x] T12.1 Windows 托盘生命周期：窗口关闭时隐藏主窗口并保持同步；托盘菜单可恢复窗口、完全退出；退出路径不被关闭钩子拦截。托盘未就绪及其他平台保留可恢复的任务栏最小化行为。验证：Windows `CGO_ENABLED=0` 下 go test / go build / go vet 通过。提交：`feat: 增加 Windows 托盘驻留与窗口恢复`。
- [x] T12.2 前端窗口行为与文案：右上角关闭按钮走统一后台关闭逻辑；设置项从“最小化到任务栏”调整为“隐藏到系统托盘”，兼容既有设置值；启动时立即同步关闭偏好。README 同步修订。验证：go test、vue-tsc 与 vite build 通过。提交：`fix: 统一关闭按钮与托盘设置行为`。
- [x] T12.3 整体验证：`go test -count=1 ./...`、`go vet ./...`、`npm run build`（vue-tsc + Vite）通过；Windows `CGO_ENABLED=0` + `desktop,production` 构建通过，验收版位于 `server/build/bin/NextPaste-tray-acceptance.exe`。本机实测普通启动与 `-minimized` 启动：托盘窗口存在、菜单恢复主窗口、原生关闭后主窗口隐藏但进程继续、再次恢复、托盘退出后进程结束，全部通过。Linux / macOS 无 GUI 标签的交叉编译通过；两平台生产 GUI 标签需 Wails 自身的 CGO 环境，本轮未配置。已清理本轮临时工具和探针。提交：`docs: 记录托盘驻留整体验证与验收程序`。

## 15. 顶部配置布局优化（2026-10-10）

- [x] T11.1 配置表单横向排列：监听地址 / 端口 / 启停按钮同排；客户端地址 / 连接按钮同排，模板区保留。窄屏按钮另起一行。验证：vue-tsc 与 vite build 通过；提交：`style: 优化顶部服务器与客户端配置表单布局`。
- [x] T11.2 顶部工具栏与连接信息对齐：模式切换 / 状态 / 收起按钮同排，下方配置与连接地址两区自适应排列，地址操作固定在右侧。860px 以下改为上下排列；连接列表限高 112px；压缩空状态，补充图标按钮名称。验证：vue-tsc / vite build、浏览器交互验收通过；提交：`style: 整理顶部工具栏与连接信息布局`。
- [x] T11.3 整体验证：最终 `npm run build`（vue-tsc + vite）通过；对构建产物进行浏览器验收，900 / 1200 / 1500px 配置与连接卡片顶端、表单控件底端对齐且顶部无横向溢出，760px 自动上下排列；深浅主题、启停 / 连接 / 断开、输入禁用、固定筛选及重载持久化、复制目标、二维码、折叠展开及重载持久化均通过，浏览器运行错误 0。保留 `output/playwright/top-layout-{light,dark}.png` 供验收，清理临时脚本和浏览器日志。

提交记录：T11.1 `bd67b20`；T11.2 `4f61647`；T11.3 `docs: 记录顶部布局整体验证与验收预览`。

测试边界：浏览器使用模拟 Wails 接口提供连接状态与网卡地址，验证了真实 Vue 组件、表单事件和地址操作；本轮仅构建前端，未重新生成桌面 exe，实际 WebView2 窗口与手机连接待用户验收。构建仍有既有的大包体提示，不影响通过。

本轮开始时已有未提交的 `auto-imports.d.ts` 与 `wailsjs/*` 生成文件修改，保留且不纳入本轮提交。

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
- [x] T10.2 工作模式区移至顶部条（可折叠），主体回归应用日志 — commit: feat: 工作模式区移至顶部
  - `AppInner.vue`：删掉 `n-layout has-sider` 左栏，改为「顶部条 + 主体」flex 列；条内并排 工作模式 tabs / 配置表单 / 连接信息，`np:workBand` 持久化收起态，收起后只留 tabs + 一行状态摘要
  - `ConnectionInfoCard.vue`：新增 `listMaxHeight`（顶部条内压到 150px）
  - 顺带修真 bug：地址行用 `n-code` 但未配置 hljs，控制台报 `[naive/code]: hljs is not set.` 且**地址文字根本不渲染**（只剩三个操作图标）→ 改自带样式的等宽文本
  - 验证：vue-tsc 通过；浏览器 dev 预览实测（展开/收起、地址三行可见、日志满宽）通过
- [x] T10.3 第三列「剪切板内容」面板：条目默认隐藏内容（文本打码、图片遮罩），单条眼睛按钮揭示 + 全局显示开关；日志表移除「剪贴板内容」列 — commit: feat: 剪切板内容独立第三列（默认隐藏）
  - 新增 `ClipboardContentPanel.vue`：条目 = clipboard 分类且后端有暂存原文的日志；隐藏态只显示类型 / 时间 / 事件描述 + 遮罩块（图片额外带体积，文本不泄漏任何字形），点眼睛或遮罩单条揭示，头部「显示内容」开关管全局（`np:clipReveal` 持久化），行内眼睛 + 复制按钮常显
  - 原文按需拉取：隐藏态完全不请求 `GetLogContent`，避免 base64 常驻内存与截图泄漏
  - `LogViewerPanel.vue`：删掉「剪贴板内容」列及其图片懒加载 / 复制代码（内容展示职责移交第三列），日志回归 级别 / 分类 / 内容 / 时间 四列事件流水
  - 验证：vue-tsc 通过；浏览器实测 5 条（3 文本 2 图片）默认全遮罩 → 单条揭示取到全文并保留换行 → 全局开关一次性揭示、图片受面板宽度约束不溢出
- [x] T10.4 条目尺寸自定义：文本框宽/高、图片宽/高四项可调（滑块 + 数字输入），第三列宽度可拖，配置持久化 `np:clipSizes`，一键恢复默认 — commit: feat: 剪切板条目尺寸可调
  - 新增 `utils/clipLayout.ts`：默认值 + 范围钳制 + localStorage 读写（尺寸 `np:clipSizes`、面板宽 `np:clipPanelWidth`），脏值/越界一律回落到边界
  - 面板头部「尺寸」弹层：四行滑块 + 数字框双向联动，恢复默认一键回到 320×140 / 200×140
  - 图片按"盒子"理解：给定宽高内 `object-fit: contain` 居中，`max-width:100%` 保证永不溢出列宽；遮罩块套同一尺寸，揭示前后布局不跳
  - 第三列左缘拖拽调宽（260–760px，双击恢复），小窗口下再按 `窗口宽-660` 上限回退，避免把日志挤没
  - 验证：vue-tsc 通过；浏览器实测改图片宽 420 → 渲染被列宽夹到 326.4px 且写入 `np:clipSizes`；拖拽列宽 380→510 持久化；恢复默认回到 200px
- [x] T10.5 图片点击放大预览 — commit: feat: 剪切板图片点击放大预览
  - 先手写了 modal 查看器，实测发现本仓库锁定的 naive-ui 2.43 `n-modal` **完全不处理 Esc**（modal 源码里没有任何 keydown，全库只有 `image/ImagePreview.mjs` 等 7 个组件监听 Escape），手写的退出路径只能靠点按钮 → 改用组件库自带 `n-image` 预览：滚轮缩放、拖拽平移、旋转/翻转、工具栏、Esc 关闭全部现成
  - 缩略图改为 `n-image`（宽高仍走 T10.4 的自定义尺寸，`object-fit: contain` + `max-width:100%`），条目下方常驻一行 `720 × 420 px · 41.0 KB · 点击可放大`（原始像素尺寸在图片加载后测得，体积取后端暂存的字节数）
  - 验证：vue-tsc + vite build 通过；无头 Chrome 实测缩略图尺寸/信息行渲染正确。放大预览的"点开→缩放→Esc 关闭"交互链路在沙箱内无法驱动（外部浏览器连接中断、无头模式过渡动画不落地），列入真机验收 T11 清单第 18 项
- [x] T10.6 整体测试与文档 — commit: test/docs: 本轮整体校验与真机验收清单
  - 校验门全绿：`go build` / `go vet` / `go test -count=1`（CGO_ENABLED=0）、darwin+linux 交叉编译、`npm run build`（vue-tsc + vite）、`wails build` 产出 12.5MB exe（27.3s）
  - README 桌面端特性段重写（三栏工作区 / 默认隐藏 / 尺寸自定义 / 放大预览 / 去重与回声抑制），fork 归属行同步
  - 沙箱内 GUI 冒烟未通过环境关：从 `build/bin` 直接起 exe 时 WebView2 报"无法读取和写入其数据目录 `…\NextPaste Server.exe\EBWebView`"（把 exe 路径当目录用），应用窗口起不来 → 端到端交互全部转入第 9 节验收清单 15–22 项
  - 外部浏览器连接（Qoder Browser Connector）中途断开，放大预览的打开/关闭链路改由无头 Chrome + 源码级核对（`n-image` 渲染出 `data-preview-src`、`ImagePreview.mjs` 监听 Escape）双重确认

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
| 2026-10-10 | T10.2 | 布局重构：删掉 `n-layout has-sider` 左栏，工作模式 / 配置表单 / 连接信息并排进顶部条，可收起为一行状态摘要（`np:workBand`），主体让位给日志；`ConnectionInfoCard` 加 `listMaxHeight`（条内 150px）。顺带修真 bug：地址行用 `n-code` 但未配 hljs → 控制台 `[naive/code]: hljs is not set.` 且**地址文字完全不渲染**（只剩三个图标），改为自带样式的等宽文本 | frontend/{AppInner.vue、components/ConnectionInfoCard.vue} | vue-tsc 通过；浏览器实测展开/收起、三条 ws:// 地址可见、日志满宽 |
| 2026-10-10 | T10.3 | 第三列剪切板面板（默认隐藏）：新增 `ClipboardContentPanel.vue`，条目 = clipboard 分类且有暂存原文的日志；遮罩态文本不露字形、图片只显体积，单条眼睛 + 全局开关（`np:clipReveal`）；原文按需拉取，隐藏态完全不请求。日志表删掉「剪贴板内容」列及其图片懒加载/复制代码，回归四列事件流水 | frontend/{components/ClipboardContentPanel.vue、AppInner.vue、components/LogViewerPanel.vue、types/index.ts、components.d.ts} | vue-tsc 通过；浏览器实测 5 条（3 文本 2 图）默认全遮罩 → 单条揭示拿到全文并保留换行 → 全局开关一次揭示且图片不溢出列宽 |
| 2026-10-10 | T10.4 | 条目尺寸自定义：新增 `utils/clipLayout.ts`（默认值 + 范围钳制 + localStorage），面板头部「尺寸」弹层四行滑块与数字框联动控制文本框宽高 / 图片宽高，恢复默认一键回 320×140 / 200×140；图片按盒子理解（contain 居中 + max-width 100% 不溢出），遮罩套同一尺寸避免揭示时跳动；第三列左缘支持拖拽调宽（260–760，双击恢复），窄窗口按"窗口宽-660"回退 | frontend/{utils/clipLayout.ts、components/ClipboardContentPanel.vue、AppInner.vue} | vue-tsc 通过；浏览器实测图片宽 420 → 渲染被列宽夹到 326.4px 且落盘、拖拽列宽 380→510 持久化、恢复默认回到 200px |
| 2026-10-10 | T10.5 | 图片放大预览：先手写 modal 查看器，实测发现本仓库 naive-ui 2.43 的 `n-modal` 源码里**没有任何 keydown**（全库仅 ImagePreview 等 7 个组件处理 Escape）→ 手写的查看器没有键盘退出，遂改用组件库 `n-image` 自带预览（滚轮缩放 / 拖拽平移 / 旋转翻转 / 工具栏 / Esc 关闭全现成）；缩略图仍走 T10.4 尺寸，条目下常驻"原始像素尺寸 · 体积 · 点击可放大" | frontend/components/ClipboardContentPanel.vue | vue-tsc + vite build 通过；无头 Chrome 确认缩略图与信息行渲染、img 带 `data-preview-src` 且未加 preview-disabled；打开/关闭交互列入真机验收第 19 项 |
| 2026-10-10 | T10.6 | 整体测试与文档：全量校验门 + `wails build` 产物 + README 桌面端特性段重写（三栏工作区 / 默认隐藏 / 尺寸自定义 / 放大预览 / 去重与回声抑制）+ 验收清单 15–22 项；沙箱内从 `build/bin` 起 GUI 失败（WebView2 报无法读写 `…\NextPaste Server.exe\EBWebView`，把 exe 当目录），端到端交互全部转入真机验收 | README.md、todo.md | go build/vet/test（-count=1）、darwin+linux 交叉编译、npm run build、wails build（27.3s / 12.5MB）全部通过 |
| 2026-10-10 | T18.x | 品牌图标（根目录 logo.png 提取图形标识：白底键出 + 反预乘 + 紧致裁剪）接入 exe/任务栏/托盘图标与标题栏；记录本地持久化（`%AppData%\NextPaste\logs.json`，原子写、启动加载、800ms 节流落盘、退出前 flush）与 `maxLogs` 可配置（50–10000，缩小即裁剪）；`GetServerClients` 暴露设备明细，表头/顶部条改为"本机作为服务器 · 已连接 N 台设备"+悬浮明细；日志表第 3 列由"内容"改为"剪贴板内容"（无预览回退事件描述）；顺带修状态刷新按序化避免首屏设备明细为空 | server/logstore.go、logstore_test.go、app.go、settings.go、build/appicon.png、build/windows/icon.ico、frontend/src/{AppInner.vue、components/{LogViewerPanel.vue、SettingsModal.vue}、assets/images/logo-mark.png}、wailsjs/go/main/{App.d.ts,App.js} | go build/vet/test（-count=1，7 项新单测 PASS）、darwin+linux 交叉编译、npm run build、wails build（20.8s / 12.8MB）、新 exe 图标 7/7 字节一致、无头 Chrome+CDP 11/11 PASS |

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
| 14 | 拖动窗口边缘改变宽度（1500 / 1200 / 1100 / 900 附近） | ~~侧栏断点~~ 已被本轮需求十一取代：现在校验顶部条是否随宽度换行、日志列是否切紧凑宽度、第三列是否被限制在"窗口宽-660"以内而不挤没日志 |

### 本轮新增（2026-10-10 下午，需求十~十三）

| # | 操作 | 预期 |
|---|---|---|
| 15 | 手机复制一段文本，看 PC 日志 | 只出现**一行**"收到文本数据 … 来自 <设备>"，不再同时冒出"已接收并写入文本数据"；第三列同步出现该条且默认遮罩 |
| 16 | PC 本地复制一段文本 | 只出现一行"检测到剪贴板文本变化"，不再有"广播剪贴板数据"第二行；对端也只收到一次 |
| 17 | 手机复制 → PC 揭示该条 → 点复制回剪贴板 | 两端都不再刷屏：写回系统剪贴板触发的监听回声被吞掉，不产生新日志、不二次广播 |
| 18 | 连续两次复制**完全相同**的内容（间隔 > 3s） | 两条独立记录（去重窗口 3s，只压链路内的重复，不压用户真的重复复制） |
| 19 | 第三列点图片缩略图 | 打开大图预览：滚轮/工具栏缩放、拖拽平移、旋转翻转、Esc 或点遮罩关闭；条目上常驻"宽 × 高 px · 体积 · 点击可放大" |
| 20 | 头部"显示内容"开关 + 单条眼睛按钮 | 开关管全局、眼睛做单条反向覆盖；重启后记住全局开关状态（`np:clipReveal`），默认仍是隐藏 |
| 21 | 「尺寸」弹层调文本框宽高 / 图片宽高；拖第三列左缘；双击左缘 | 实时生效并持久化（`np:clipSizes`）；列宽持久化（`np:clipPanelWidth`）；双击回到默认 380px；恢复默认按钮回到 320×140 / 200×140 |
| 22 | 顶部条点"收起配置"后重启 | 只留模式 tabs + 一行状态摘要（`np:workBand` 记住），日志区高度变大；连接信息里地址文字可见（ws://…），图钉/二维码/复制可用 |
| 23 | 顶部布局优化验收：使用最新前端启动桌面开发版或重新构建 exe，切换服务器 / 客户端，在 900 / 1200 / 1500px 宽度检查顶部并切换深浅主题 | 模式、状态、收起按钮同一行；监听地址、端口与启停按钮同排；配置与连接区顶端对齐，地址操作靠右；收起后日志区高度增加；运行时输入禁用但停止 / 断开按钮可用；二维码扫码及实际连接正常 |

### 本轮新增（2026-10-10 晚，T18.1–T18.5）

| # | 操作 | 预期 |
|---|---|---|
| 24 | 重启桌面端后看 exe / 任务栏 / 托盘图标与标题栏 | 图标为传笺标识（笔记本电脑+信笺+手机）；标题栏左侧出现标识，右侧文案与 NPBP 标签不变 |
| 25 | 制造若干条同步记录后完全退出应用，再重新打开 | 记录仍在（读自 `%AppData%\NextPaste\logs.json`）；设置面板「存储位置」显示该路径；点「清空」后重启不再回来 |
| 26 | 设置面板把「记录数量」改小（如 100）并观察 | 立即只保留最新 100 条（更早的记录被删除）；重启后仍是 100 条上限 |
| 27 | 服务器模式下手机连接 | 表头显示「已连接 N 台设备」，悬浮可看设备名 / 平台 / 连接时间；顶部条摘要为「本机作为服务器 · 端口 X · 已连接 N 台设备」；日志表第 3 列为「剪贴板内容」，文本记录显示内容预览、连接/系统记录显示事件描述 |
