# 桌面快捷方式与真实托盘点击修复（2026-10-11）

## 原因与修复

桌面 `传笺.lnk` 带 `-minimized`，第一次启动会隐藏到托盘，看起来没有打开窗口。本次仅将桌面快捷方式参数清空；开机自启注册项仍保留 `-minimized -autostart=server`。

托盘旧实现只在 `GetMessage` 返回后拦截消息。直接送到窗口过程的通知不会进入这段分支，落入 systray 的默认空菜单；真实菜单命令同样可能绕过应用处理。Windows 的直接消息与队列消息确实有不同的派发路径，见 [GetMessageW 官方说明](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getmessagew)。之前仅使用 PostMessage 验收，因此“真实点击已验证”的结论不充分。

本次在托盘窗口所属线程安装 `SetWindowSubclass`，统一处理直接通知和队列通知，未处理的消息继续交给原托盘库；安装完成后才标记就绪。窗口销毁时移除回调，保留消息循环固定线程和原生菜单行为。

## 验证

- 在旧代码上新增 `SendMessageTimeoutW` 左键通知测试，明确复现回调等待超时。
- 修复后直接左键、直接右键弹出/取消、直接设置命令通过。
- 原有 25 次队列左键、队列右键、菜单命令分流、菜单重复打开/取消及托盘退出回归通过。
- 全量 `go test -count=1 ./...`、`go vet ./...`、Wails v2.11.0 生产构建通过。
- 使用 computer-use 在正式窗口打开设置并正常退出旧程序；没有强制终止进程。
- 保留旧版 `server/build/bin/ChuanJian.pre-sent-message-F392F2B4.exe` 后更新正式 exe，并正常显示启动。
- **用户物理验收已确认：“菜单能弹出，打开设置也正常”。**

正式程序：`server/build/bin/ChuanJian.exe`。
SHA256：`D90A4F0C85AF24AC3D44B07B003C5784E75E90C7A361147C733539D17806ACDE`。
桌面入口：OneDrive 桌面的 `传笺.lnk`，目标、工作目录与绿色图标保持不变，参数改为空。

源码修复提交：`e868ec7`；交付记录单独提交。临时构建 exe 和工具清理，永久回归测试与旧版备份保留，用户 `logo.png` 未改动，未推送。

手机同步、HarmonyOS 构建和实际重启自启未重复验证。用户已确认右键菜单和设置，其他菜单动作仍保留自动消息测试证据，不宣称每项均做过物理点击。
