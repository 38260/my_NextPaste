<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  NButton,
  NCard,
  NCheckbox,
  NEmpty,
  NIcon,
  NList,
  NListItem,
  NModal,
  NScrollbar,
  NSpace,
  NText,
  NTooltip,
  useMessage
} from 'naive-ui'
import { GetLocalIPs } from '../../wailsjs/go/main/App'
import QRCode from 'qrcode'
import { CopyOutline, PinOutline, ScanOutline } from '@vicons/ionicons5'

interface Props {
  isRunning: boolean
  port?: number
  mode?: 'server' | 'client'
  serverUrl?: string
  listMaxHeight?: number
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'server',
  serverUrl: '',
  port: 8080,
  listMaxHeight: 220,
  compact: false
})

const message = useMessage()

const addresses = ref<string[]>([])

const showQrModal = ref(false)
const qrCodeUrl = ref('')
const currentQrAddress = ref('')

const title = computed(() => (props.mode === 'client' ? '服务器地址' : '连接信息'))

const PINNED_KEY = 'np:pinnedAddresses'

const pinnedAddresses = ref<string[]>(loadPinned())
const onlyPinned = ref(false)

function loadPinned(): string[] {
  try {
    const raw = localStorage.getItem(PINNED_KEY)
    const arr = raw ? JSON.parse(raw) : []
    return Array.isArray(arr) ? arr.filter((x) => typeof x === 'string') : []
  } catch {
    return []
  }
}

function savePinned(list: string[]) {
  pinnedAddresses.value = list
  localStorage.setItem(PINNED_KEY, JSON.stringify(list))
}

const isPinned = (addr: string) => pinnedAddresses.value.includes(addr)

const togglePin = (addr: string) => {
  if (isPinned(addr)) {
    savePinned(pinnedAddresses.value.filter((a) => a !== addr))
    message.info('已取消固定该地址')
  } else {
    savePinned([...pinnedAddresses.value, addr])
    message.success('已固定，重启后仍置顶显示')
  }
}

interface DisplayAddress {
  addr: string
  pinned: boolean
  online: boolean
}

// 固定地址（pin 顺序）置顶；枚举不到的固定地址保留显示并灰态标注
const displayAddresses = computed<DisplayAddress[]>(() => {
  const live = addresses.value
  const pinnedItems: DisplayAddress[] = pinnedAddresses.value.map((addr) => ({
    addr,
    pinned: true,
    online: live.includes(addr)
  }))
  if (onlyPinned.value) {
    return pinnedItems
  }
  const liveItems: DisplayAddress[] = live
    .filter((addr) => !isPinned(addr))
    .map((addr) => ({ addr, pinned: false, online: true }))
  return [...pinnedItems, ...liveItems]
})

const loadAddresses = async () => {
  if (props.mode === 'client') {
    addresses.value = props.serverUrl ? [props.serverUrl] : []
    return
  }

  if (!props.isRunning) {
    addresses.value = []
    return
  }

  try {
    const ips = await GetLocalIPs()
    addresses.value = ips.map((ip: string) => `ws://${ip}:${props.port}/ws`)
  } catch {
    addresses.value = []
  }
}

const copyAddress = async (addr: string) => {
  try {
    await navigator.clipboard.writeText(addr)
    message.success('已复制')
  } catch {
    message.error('复制失败')
  }
}

const showQrCode = async (addr: string) => {
  currentQrAddress.value = addr
  showQrModal.value = true
  qrCodeUrl.value = ''
  try {
    qrCodeUrl.value = await QRCode.toDataURL(addr, {
      margin: 2,
      width: 220,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    })
  } catch {
    qrCodeUrl.value = ''
  }
}

watch(
  () => [props.isRunning, props.port, props.serverUrl, props.mode],
  () => {
    loadAddresses()
  }
)

onMounted(() => {
  if (props.isRunning || (props.mode === 'client' && props.serverUrl)) {
    loadAddresses()
  }
})
</script>

<template>
  <n-card size="small" :title="title" :class="{ 'connection-compact': compact }">
    <template v-if="mode === 'server'" #header-extra>
      <n-checkbox v-model:checked="onlyPinned" size="small">
        <n-text depth="3" style="font-size: 12px">仅显示固定地址</n-text>
      </n-checkbox>
    </template>

    <template v-if="!isRunning && mode === 'server'">
      <n-empty description="服务器未运行" />
    </template>

    <template v-else-if="!isRunning && mode === 'client'">
      <n-empty description="未连接服务器" />
    </template>

    <template v-else>
      <n-space vertical :size="compact ? 8 : 12">
        <div>
          <n-text depth="3" :style="compact ? 'font-size: 12px' : undefined">{{ mode === 'client' ? '当前连接地址' : '可用连接地址' }}</n-text>
        </div>

        <n-scrollbar :style="`max-height: ${listMaxHeight}px`">
          <n-list v-if="displayAddresses.length" bordered>
            <n-list-item v-for="item in displayAddresses" :key="item.addr">
              <div class="address-row">
                <div class="address-value">
                  <span class="addr-text" :style="item.online ? '' : 'opacity: 0.4'">
                    {{ item.addr }}
                  </span>
                  <div v-if="!item.online" style="margin-top: 2px">
                    <n-text depth="3" style="font-size: 12px">当前网络不在此网段</n-text>
                  </div>
                </div>
                <n-space :size="4" class="address-actions" :wrap="false">
                  <n-tooltip v-if="mode === 'server'" trigger="hover">
                    <template #trigger>
                      <n-button
                        size="tiny"
                        quaternary
                        :aria-label="item.pinned ? '取消固定地址' : '固定地址'"
                        :type="item.pinned ? 'primary' : 'default'"
                        @click="togglePin(item.addr)"
                      >
                        <template #icon>
                          <n-icon><PinOutline /></n-icon>
                        </template>
                      </n-button>
                    </template>
                    {{ item.pinned ? '取消固定' : '固定此地址' }}
                  </n-tooltip>

                  <n-tooltip trigger="hover">
                    <template #trigger>
                      <n-button size="tiny" quaternary aria-label="显示连接二维码" @click="showQrCode(item.addr)">
                        <template #icon>
                          <n-icon><ScanOutline /></n-icon>
                        </template>
                      </n-button>
                    </template>
                    二维码
                  </n-tooltip>

                  <n-tooltip trigger="hover">
                    <template #trigger>
                      <n-button size="tiny" quaternary type="primary" aria-label="复制连接地址" @click="copyAddress(item.addr)">
                        <template #icon>
                          <n-icon><CopyOutline /></n-icon>
                        </template>
                      </n-button>
                    </template>
                    复制
                  </n-tooltip>
                </n-space>
              </div>
            </n-list-item>
          </n-list>

          <n-empty
            v-if="!displayAddresses.length"
            :description="onlyPinned ? '暂无固定地址，点击地址旁的图钉固定' : '正在获取网络地址...'"
          />
        </n-scrollbar>
      </n-space>
    </template>

    <n-modal v-model:show="showQrModal" preset="card" title="扫码连接" style="width: 360px">
      <n-space vertical align="center" :size="12">
        <img v-if="qrCodeUrl" :src="qrCodeUrl" alt="QR Code" style="width: 220px; height: 220px" />
        <n-empty v-else description="生成中..." />
      </n-space>
    </n-modal>
  </n-card>
</template>

<style scoped>
.address-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.address-value {
  flex: 1;
  min-width: 0;
}

.address-actions {
  flex-shrink: 0;
}

.connection-compact :deep(.n-list-item) {
  padding: 6px 10px;
}

.connection-compact :deep(.n-empty) {
  flex-direction: row;
  justify-content: flex-start;
  align-items: center;
  gap: 8px;
  min-height: 60px;
}

.connection-compact :deep(.n-empty__icon) {
  margin: 0;
  width: 22px;
  height: 22px;
  font-size: 22px;
}

.connection-compact :deep(.n-empty__description) {
  margin: 0;
}

/* n-code 需要额外配置 hljs，未配置时地址文字根本不渲染，这里改为自带样式的等宽文本 */
.addr-text {
  display: inline-block;
  font-family: ui-monospace, SFMono-Regular, Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(128, 128, 128, 0.14);
  word-break: break-all;
}
</style>
