<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, watchEffect } from 'vue'
import {
  EventsOff,
  EventsOn,
  WindowMinimise
} from '../wailsjs/runtime/runtime'
import {
  ApplySettings,
  ClearLogs,
  ConnectClient,
  DisconnectClient,
  GetAutostartStatus,
  GetClientStatus,
  GetLogs,
  GetServerStatus,
  HideWindow,
  Quit,
  StartServer,
  StopServer
} from '../wailsjs/go/main/App'
import type { ClipItem, LogEntry, ServerConfig as ServerConfigType, ServerStatus } from './types'
import { useDialog, useMessage } from 'naive-ui'

import ServerConfigForm from './components/ServerConfigForm.vue'
import ClientConfigForm from './components/ClientConfigForm.vue'
import ConnectionInfoCard from './components/ConnectionInfoCard.vue'
import LogViewerPanel from './components/LogViewerPanel.vue'
import ClipboardContentPanel from './components/ClipboardContentPanel.vue'
import StatusIndicatorTag from './components/StatusIndicatorTag.vue'
import SettingsModal from './components/SettingsModal.vue'

import {
  DEFAULT_PANEL_WIDTH,
  clampPanelWidth,
  loadPanelWidth,
  savePanelWidth
} from './utils/clipLayout'

import {
  CopyOutline,
  CloseOutline,
  ColorPaletteOutline,
  SettingsOutline as CogOutline,
  MoonOutline,
  RemoveOutline,
  ScanOutline,
  SunnyOutline
} from '@vicons/ionicons5'

type Mode = 'server' | 'client'

type ColorScheme = 'light' | 'dark' | 'auto'

interface Props {
  colorScheme?: ColorScheme
}

interface Emits {
  (e: 'update:colorScheme', value: ColorScheme): void
}

const props = withDefaults(defineProps<Props>(), {
  colorScheme: 'auto'
})

const emit = defineEmits<Emits>()

const message = useMessage()
const dialog = useDialog()

const mode = ref<Mode>('server')

const config = ref<ServerConfigType>({
  address: '0.0.0.0',
  port: 8080
})

const clientUrl = ref('ws://localhost:8080/ws/my-room')

const status = ref<ServerStatus>({
  isRunning: false,
  clientCount: 0
})

const clientStatus = ref({
  isConnected: false,
  serverUrl: ''
})

const logs = ref<LogEntry[]>([])

const hasActiveConnection = computed(() => {
  return mode.value === 'server' ? status.value.isRunning : clientStatus.value.isConnected
})

const loadConfig = () => {
  const savedMode = localStorage.getItem('mode')
  if (savedMode === 'server' || savedMode === 'client') {
    mode.value = savedMode
  }

  const saved = localStorage.getItem('serverConfig')
  if (saved) {
    try {
      config.value = JSON.parse(saved)
    } catch {
      // ignore
    }
  }

  const savedClientUrl = localStorage.getItem('clientUrl')
  if (savedClientUrl) {
    clientUrl.value = savedClientUrl
  }
}

const saveConfig = () => {
  localStorage.setItem('serverConfig', JSON.stringify(config.value))
}

const saveClientUrl = () => {
  localStorage.setItem('clientUrl', clientUrl.value)
}

const saveMode = () => {
  localStorage.setItem('mode', mode.value)
}

const switchMode = (newMode: Mode) => {
  if (hasActiveConnection.value) {
    message.warning('请先停止当前连接')
    return
  }
  mode.value = newMode
  saveMode()
}

const handleStart = async (cfg: ServerConfigType) => {
  try {
    await StartServer(cfg.address, cfg.port)
    config.value = cfg
    saveConfig()
    await updateStatus()
    message.success('服务已启动')
  } catch (error) {
    dialog.error({
      title: '启动失败',
      content: String(error)
    })
  }
}

const handleStop = async () => {
  try {
    await StopServer()
    await updateStatus()
    message.info('服务已停止')
  } catch (error) {
    dialog.error({
      title: '停止失败',
      content: String(error)
    })
  }
}

const updateStatus = async () => {
  try {
    const newStatus = await GetServerStatus()
    status.value = newStatus as ServerStatus
  } catch {
    // ignore
  }
}

const loadLogs = async () => {
  try {
    const logList = await GetLogs()
    logs.value = logList as LogEntry[]
  } catch {
    // ignore
  }
}

const handleClearLogs = async () => {
  try {
    await ClearLogs()
    logs.value = []
    message.success('已清空日志')
  } catch {
    // ignore
  }
}

const handleConnect = async (url: string) => {
  try {
    await ConnectClient(url)
    clientUrl.value = url
    saveClientUrl()
    await updateClientStatus()
    message.success('已连接')
  } catch (error) {
    dialog.error({
      title: '连接失败',
      content: String(error)
    })
  }
}

const handleDisconnect = async () => {
  try {
    await DisconnectClient()
    await updateClientStatus()
    message.info('已断开')
  } catch (error) {
    dialog.error({
      title: '断开失败',
      content: String(error)
    })
  }
}

const updateClientStatus = async () => {
  try {
    const newStatus = await GetClientStatus()
    clientStatus.value = newStatus as any
  } catch {
    // ignore
  }
}

const onLogsUpdated = (newLogs: LogEntry[]) => {
  logs.value = newLogs
}

const handleMinimize = () => {
  WindowMinimise()
}

const handleCloseOrQuit = async () => {
  if (settings.value.closeAction === 'close') {
    await Quit()
    return
  }
  await HideWindow()
}

const closeTip = computed(() =>
  settings.value.closeAction === 'close' ? '退出应用' : '隐藏到系统托盘'
)

onMounted(() => {
  loadConfig()
  updateStatus()
  updateClientStatus()
  loadLogs()

  EventsOn('logs:updated', onLogsUpdated)

  const statusInterval = window.setInterval(() => {
    updateStatus()
    updateClientStatus()
  }, 2000)

  onUnmounted(() => {
    window.clearInterval(statusInterval)
    EventsOff('logs:updated')
  })
})

const colorScheme = computed<ColorScheme>({
  get: () => props.colorScheme,
  set: (v) => emit('update:colorScheme', v)
})

const settingsOpen = ref(false)

// 响应式布局：窗口宽度驱动日志表列宽与顶部条形态
const windowWidth = ref(window.innerWidth)
const onWindowResize = () => {
  windowWidth.value = window.innerWidth
}

// 第三列宽度：用户可拖动调节并持久化；未调过时按窗口宽度给默认值
const clipWidth = ref(loadPanelWidth())

const effClipWidth = computed(() =>
  Math.min(clipWidth.value, Math.max(260, windowWidth.value - 660))
)

const compactMode = computed(() => windowWidth.value - effClipWidth.value < 760)

const clipDragging = ref(false)

const onClipDragMove = (e: MouseEvent) => {
  if (!clipDragging.value) return
  clipWidth.value = clampPanelWidth(window.innerWidth - 16 - e.clientX)
}

const stopClipDrag = () => {
  clipDragging.value = false
  savePanelWidth(clipWidth.value)
  window.removeEventListener('mousemove', onClipDragMove)
  window.removeEventListener('mouseup', stopClipDrag)
}

const startClipDrag = () => {
  clipDragging.value = true
  window.addEventListener('mousemove', onClipDragMove)
  window.addEventListener('mouseup', stopClipDrag)
}

const resetClipWidth = () => {
  clipWidth.value = DEFAULT_PANEL_WIDTH
  savePanelWidth(DEFAULT_PANEL_WIDTH)
}

// 剪切板条目 = clipboard 分类且后端暂存了原文的日志
const clipItems = computed<ClipItem[]>(() =>
  logs.value
    .filter((l) => l.category === 'clipboard' && l.contentId)
    .map((l) => ({
      id: l.contentId as number,
      kind: (l.preview && l.preview.startsWith('图片') ? 'image' : 'text') as 'text' | 'image',
      caption: l.message,
      preview: l.preview ?? '',
      timestamp: l.timestamp
    }))
)

// 顶部工作模式条：展开时含配置与连接信息，收起时只留模式切换
const BAND_KEY = 'np:workBand'
const bandCollapsed = ref(localStorage.getItem(BAND_KEY) === 'collapsed')

const toggleBand = () => {
  bandCollapsed.value = !bandCollapsed.value
  localStorage.setItem(BAND_KEY, bandCollapsed.value ? 'collapsed' : 'expanded')
}

const bandSummary = computed(() => {
  if (mode.value === 'server') {
    return status.value.isRunning
      ? `服务器运行中 · 端口 ${config.value.port} · 客户端 ${status.value.clientCount}`
      : '服务器未启动'
  }
  return clientStatus.value.isConnected ? '已连接服务器' : '客户端未连接'
})

onMounted(() => {
  window.addEventListener('resize', onWindowResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onWindowResize)
})

type AutoStartMode = 'server' | 'client'

type Settings = {
  autoStartEnabled: boolean
  autoStartMode: AutoStartMode
  autoStartActionEnabled: boolean
  closeAction: 'tray' | 'close'
}

const SETTINGS_KEY = 'np:settings'

const settings = ref<Settings>({
  autoStartEnabled: false,
  autoStartMode: 'server',
  autoStartActionEnabled: false,
  closeAction: 'tray'
})

watchEffect(() => {
  const raw = localStorage.getItem(SETTINGS_KEY)
  if (!raw) return
  try {
    const obj = JSON.parse(raw) as Partial<Settings>
    settings.value = {
      autoStartEnabled: Boolean(obj.autoStartEnabled),
      autoStartMode: obj.autoStartMode === 'client' ? 'client' : 'server',
      autoStartActionEnabled: Boolean(obj.autoStartActionEnabled),
      closeAction: obj.closeAction === 'close' ? 'close' : 'tray'
    }
  } catch {
    // ignore
  }
})

const didAutoAction = ref(false)

onMounted(async () => {
  if (didAutoAction.value) return
  if (!settings.value.autoStartActionEnabled) return

  didAutoAction.value = true

  try {
    if (settings.value.autoStartMode === 'server') {
      mode.value = 'server'
      await StartServer(config.value.address, config.value.port)
      await updateStatus()
      message.success('已自动启动服务端')
    } else {
      mode.value = 'client'
      await ConnectClient(clientUrl.value)
      await updateClientStatus()
      message.success('已自动连接')
    }
  } catch (error) {
    dialog.error({
      title: '自动执行失败',
      content: String(error)
    })
  }
})

watchEffect(() => {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings.value))
})

const refreshAutostartStatus = async () => {
  try {
    const st = (await GetAutostartStatus()) as {
      enabled: boolean
      mode: string
      error: string
    }
    if (st.error) return
    settings.value.autoStartEnabled = st.enabled
    if (st.mode === 'client' || st.mode === 'server') {
      settings.value.autoStartMode = st.mode
    }
  } catch {
    // ignore
  }
}

let applySettingsTimer: number | undefined

watch(
  settings,
  (_current, previous) => {
    window.clearTimeout(applySettingsTimer)
    const apply = async () => {
      try {
        await ApplySettings(JSON.stringify(settings.value))
      } catch (error) {
        message.error(`开机自启设置未生效：${String(error)}`)
        await refreshAutostartStatus()
      }
    }
    if (previous === undefined) {
      void apply()
    } else {
      applySettingsTimer = window.setTimeout(apply, 300)
    }
  },
  { deep: true, immediate: true }
)
</script>

<template>
  <n-layout embedded style="height: 100vh">
    <n-layout-header bordered style="height: 56px">
      <div class="header">
        <div class="header-drag" />
        <div class="header-left">
          <n-space align="center" :size="12">
            <n-text strong>NextPaste Server</n-text>
            <n-tag size="small" type="info" round>NPBP V1.1</n-tag>
          </n-space>
        </div>

        <div class="header-right">
          <n-space align="center" :size="12">
            <StatusIndicatorTag :is-running="hasActiveConnection" />

            <template v-if="mode === 'server' && status.isRunning">
              <n-tag size="small" round type="success">客户端 {{ status.clientCount }}</n-tag>
            </template>

            <n-radio-group v-model:value="colorScheme" size="small" name="theme">
              <n-radio-button value="auto">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-icon><ColorPaletteOutline /></n-icon>
                  </template>
                  跟随系统
                </n-tooltip>
              </n-radio-button>
              <n-radio-button value="light">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-icon><SunnyOutline /></n-icon>
                  </template>
                  浅色
                </n-tooltip>
              </n-radio-button>
              <n-radio-button value="dark">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-icon><MoonOutline /></n-icon>
                  </template>
                  深色
                </n-tooltip>
              </n-radio-button>
            </n-radio-group>

            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary size="small" @click="settingsOpen = true">
                  <template #icon><n-icon><CogOutline /></n-icon></template>
                </n-button>
              </template>
              设置
            </n-tooltip>

            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary size="small" @click="handleMinimize">
                  <template #icon><n-icon><RemoveOutline /></n-icon></template>
                </n-button>
              </template>
              最小化
            </n-tooltip>

            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary size="small" type="error" @click="handleCloseOrQuit">
                  <template #icon><n-icon><CloseOutline /></n-icon></template>
                </n-button>
              </template>
              {{ closeTip }}
            </n-tooltip>
          </n-space>
        </div>
      </div>
    </n-layout-header>

    <div class="page">
      <div class="work-band">
        <div class="band-toolbar">
          <div class="band-mode">
            <n-text depth="3" class="band-label">工作模式</n-text>
            <n-tabs
              type="segment"
              :value="mode"
              :disabled="hasActiveConnection"
              class="band-tabs"
              @update:value="(v: string | number) => switchMode(v as any)"
            >
              <n-tab name="server">服务器</n-tab>
              <n-tab name="client">客户端</n-tab>
            </n-tabs>
          </div>

          <div class="band-summary">
            <n-text depth="3">{{ bandSummary }}</n-text>
          </div>

          <n-button
            class="band-toggle"
            size="small"
            quaternary
            :aria-expanded="!bandCollapsed"
            aria-controls="work-band-details"
            @click="toggleBand"
          >
            {{ bandCollapsed ? '展开配置' : '收起配置' }}
          </n-button>
        </div>

        <div v-if="!bandCollapsed" id="work-band-details" class="band-details">
          <div class="band-config">
            <ServerConfigForm
              v-if="mode === 'server'"
              :config="config"
              :is-running="status.isRunning"
              @start="handleStart"
              @stop="handleStop"
              @update:config="saveConfig"
            />

            <ClientConfigForm
              v-if="mode === 'client'"
              :url="clientUrl"
              :is-connected="clientStatus.isConnected"
              @connect="handleConnect"
              @disconnect="handleDisconnect"
              @update:url="saveClientUrl"
            />
          </div>

          <ConnectionInfoCard
            v-if="mode === 'server' || (mode === 'client' && clientStatus.isConnected)"
            class="band-conn"
            compact
            :list-max-height="112"
            :is-running="hasActiveConnection"
            :port="config.port"
            :mode="mode"
            :server-url="clientUrl"
          />
        </div>
      </div>

      <div class="page-body">
        <div class="body-log">
          <LogViewerPanel :logs="logs" :compact="compactMode" @clear="handleClearLogs" />
        </div>

        <div class="body-clip" :style="{ width: effClipWidth + 'px' }">
          <div
            class="clip-resizer"
            :class="{ 'is-active': clipDragging }"
            title="拖动调节宽度，双击恢复默认"
            @mousedown.prevent="startClipDrag"
            @dblclick="resetClipWidth"
          />
          <ClipboardContentPanel :items="clipItems" />
        </div>
      </div>
    </div>

    <SettingsModal
      v-model:show="settingsOpen"
      v-model:settings="settings"
    />
  </n-layout>
</template>

<style scoped>
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  position: relative;
}

.header-drag {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  right: 220px;
  --wails-draggable: drag;
}

.header-left,
.header-right {
  position: relative;
  z-index: 1;
}

.page {
  height: calc(100vh - 56px);
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.work-band {
  padding: 10px 16px 12px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.24);
}

.band-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 34px;
}

.band-mode {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.band-tabs {
  width: 180px;
}

.band-label {
  font-size: 12px;
  white-space: nowrap;
}

.band-details {
  display: grid;
  grid-template-columns: minmax(380px, 0.95fr) minmax(0, 1.05fr);
  align-items: start;
  gap: 12px;
  margin-top: 10px;
}

.band-config,
.band-conn {
  min-width: 0;
}

.band-details :deep(.n-card-header) {
  padding: 12px 16px 8px;
  font-size: 14px;
}

.band-details :deep(.n-card__content) {
  padding: 0 16px 12px;
}

.band-summary {
  min-width: 0;
  font-size: 13px;
  overflow-wrap: anywhere;
}

.band-toggle {
  margin-left: auto;
  flex-shrink: 0;
}

@media (max-width: 860px) {
  .band-details {
    grid-template-columns: minmax(0, 1fr);
  }

  .band-toolbar {
    flex-wrap: wrap;
    gap: 8px 12px;
  }

  .band-summary {
    order: 1;
    flex-basis: 100%;
  }
}

.page-body {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: stretch;
  gap: 16px;
  padding: 16px;
  overflow: hidden;
}

.body-log {
  flex: 1 1 auto;
  min-width: 0;
  overflow: auto;
}

.body-clip {
  flex: 0 0 auto;
  min-width: 260px;
  height: 100%;
  position: relative;
}

.clip-resizer {
  position: absolute;
  left: -10px;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  border-radius: 4px;
  background: transparent;
  z-index: 2;
}

.clip-resizer:hover,
.clip-resizer.is-active {
  background: rgba(24, 160, 88, 0.35);
}
</style>
