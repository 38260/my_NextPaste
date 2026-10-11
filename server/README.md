# 传笺 桌面端（ChuanJian Desktop）

传笺（ChuanJian）的 PC 桌面端应用，基于 Wails 框架开发，提供 WebSocket 服务器和剪贴板监听功能。

## 功能特性

### 🚀 核心功能

- **WebSocket 服务器**：支持多客户端连接，实现跨设备剪贴板同步
- **剪贴板监听**：实时监听系统剪贴板变化（文本和图片）
- **自动广播**：剪贴板变化自动同步到所有连接的客户端
- **协议支持**：完整实现 NextPaste 自定义协议（握手、同步、心跳）

### 💻 用户界面

- **服务器配置**：可配置监听地址和端口
- **连接信息**：自动检测并显示所有可用的 WebSocket 连接地址
- **一键复制**：快速复制连接地址到剪贴板
- **实时日志**：支持日志级别过滤和清空
- **状态监控**：实时显示服务运行状态和客户端数量
- **剪贴板内容表**：文本/图片内容在日志表内展示，默认遮罩、按需揭示，支持图片放大与一键复制

### 🖥 系统集成

- **托盘常驻**：关闭窗口隐藏到系统托盘；左键恢复主界面，右键弹出菜单——最近 3 条剪贴板记录（点击定位到日志表，跟随内容隐藏偏好）、开机自启开关、打开设置、退出应用
- **单实例**：重复启动自动唤起已有窗口，只保留一个进程和托盘图标
- **开机自启**：设置面板或托盘菜单一键开关（Windows 注册表 Run 键，纯 Go 无 CGO 依赖）
- **本地记录**：全文与图片保存在项目 `data/` 目录，随记录数量上限淘汰，重启后可读取和复制

## Windows 图片监听说明

为提升 Windows 端图片剪贴板的兼容性（尤其是微信截图、Snipaste 等常见工具），服务端使用了对 `golang.design/x/clipboard` 的 fork：

- `go.mod` 通过 `replace` 指向：`golang.design/x/clipboard => github.com/Okysu/clipboard v0.0.0-20260207041531-2b2e2d7c4eca`
- `FmtImage` 读取策略扩展：按优先级尝试
  - `CF_DIBV5` -> `CF_DIB` -> 注册格式 `PNG` -> `CF_BITMAP`（`HBITMAP`，GDI 转 PNG）
- `Watch` 修复/增强：增加 `ticker.Stop()`，读取失败仍推进 sequence number，避免重复触发
- 新增 Windows-only 测试：`clipboard_windows_watch_test.go`

效果：Windows 上微信截图和 Snipaste 等写入图片时，可以稳定触发监听并读到 PNG bytes。

## 快速开始

### 前置要求

- Go 1.24+
- Node.js 18+
- Wails CLI v2

### 安装依赖

```bash
# 安装 Go 依赖
# 项目已在 go.mod 中通过 replace 使用 github.com/Okysu/clipboard
cd ..
cd server
go mod tidy

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 开发模式

```bash
wails dev
```

### 构建应用

```bash
wails build
```

构建产物：`build/bin/ChuanJian.exe`（exe / 任务栏 / 托盘统一绿色传笺图标）。

## 使用说明

1. **启动应用**：运行构建产物 `ChuanJian.exe`，手动启动会显示主窗口
2. **配置服务器**：监听地址默认 `0.0.0.0`，端口默认 `8080`
3. **启动服务**：点击"启动服务"按钮
4. **获取连接地址**：在"连接信息"区域查看所有可用的 WebSocket 地址
5. **连接客户端**：复制连接地址到 HarmonyOS 客户端
6. **开始同步**：剪贴板内容将自动在所有设备间同步

> 关闭窗口默认隐藏到系统托盘；托盘左键恢复主界面，右键菜单提供最近剪贴板记录、开机自启开关、打开设置与退出应用。重复启动会唤起已有窗口，不会出现第二个实例。

## 技术架构

### 后端技术栈

- **Go 1.24**：主要编程语言
- **Wails v2**：桌面应用框架
- **gorilla/websocket**：WebSocket 实现
- **github.com/Okysu/clipboard**：跨平台剪贴板库（fork 自 `golang.design/x/clipboard`，Windows 端增强图片读取与监听）

### 前端技术栈

- **Vue 3**：渐进式 JavaScript 框架
- **TypeScript**：类型安全
- **Vite**：快速的前端构建工具

## 故障排除

### 端口被占用

如果默认端口 8080 被占用，请修改为其他端口（如 8081、9000 等）。

### 防火墙问题

确保 WebSocket 端口在 Windows 防火墙中开放：

```powershell
# 以管理员身份运行
netsh advfirewall firewall add rule name="ChuanJian" dir=in action=allow protocol=TCP localport=8080
```

## 开发指南

详见 [SETUP.md](./SETUP.md)

## 相关链接

- [Wails 官方文档](https://wails.io/)
- [Vue 3 文档](https://vuejs.org/)
- [NextPaste 协议文档](../docs/protocol.md)
