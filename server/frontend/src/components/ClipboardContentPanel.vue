<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  NButton,
  NCard,
  NEllipsis,
  NEmpty,
  NIcon,
  NScrollbar,
  NSpace,
  NSwitch,
  NTag,
  NText,
  useMessage
} from 'naive-ui'
import {
  CopyOutline,
  EyeOffOutline,
  EyeOutline,
  LockClosedOutline
} from '@vicons/ionicons5'
import { CopyLogContent, GetLogContent } from '../../wailsjs/go/main/App'
import type { ClipItem } from '../types'

interface Props {
  items: ClipItem[]
}

const props = defineProps<Props>()
const message = useMessage()

// 内容默认隐藏：全局开关决定所有条目的默认态，单条眼睛按钮只做反向覆盖
const MASK_KEY = 'np:clipReveal'
const revealAll = ref(localStorage.getItem(MASK_KEY) === '1')
const overrides = ref<Record<number, boolean>>({})

const isShown = (id: number) => overrides.value[id] ?? revealAll.value

const toggleAll = (v: boolean) => {
  revealAll.value = v
  overrides.value = {}
  localStorage.setItem(MASK_KEY, v ? '1' : '0')
}

const toggleOne = (id: number) => {
  overrides.value = { ...overrides.value, [id]: !isShown(id) }
}

// 内容按需拉取：隐藏态不取原文，避免无谓的 base64 常驻内存
const texts = ref<Record<number, string>>({})
const images = ref<Record<number, string>>({})
const pending = new Set<number>()

const load = (item: ClipItem) => {
  if (pending.has(item.id)) return
  if (item.kind === 'text' && texts.value[item.id] !== undefined) return
  if (item.kind === 'image' && images.value[item.id]) return

  pending.add(item.id)
  GetLogContent(item.id)
    .then((res: any) => {
      if (!res || !res.found) return
      if (res.type === 'text') {
        texts.value = { ...texts.value, [item.id]: res.text ?? '' }
      } else if (res.type === 'image' && res.imageBase64) {
        images.value = {
          ...images.value,
          [item.id]: `data:image/png;base64,${res.imageBase64}`
        }
      }
    })
    .catch(() => {})
    .finally(() => pending.delete(item.id))
}

watch(
  () => [props.items, revealAll.value, overrides.value],
  () => {
    props.items.forEach((item) => {
      if (isShown(item.id)) load(item)
    })
  },
  { deep: true, immediate: true }
)

const copy = async (item: ClipItem) => {
  try {
    await CopyLogContent(item.id)
    message.success('已复制到剪贴板')
  } catch (error) {
    message.error(`复制失败：${String(error)}`)
  }
}

const fmtTime = (ts: number) => {
  if (!ts) return '--:--:--'
  return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false })
}
</script>

<template>
  <n-card size="small" class="clip-panel">
    <template #header>
      <n-space align="baseline" :size="8">
        <n-text strong>剪切板内容</n-text>
        <n-text depth="3" style="font-size: 12px">{{ items.length }} 条</n-text>
      </n-space>
    </template>

    <template #header-extra>
      <n-space align="center" :size="8">
        <n-text depth="3" style="font-size: 12px">显示内容</n-text>
        <n-switch size="small" :value="revealAll" @update:value="toggleAll" />
      </n-space>
    </template>

    <n-empty v-if="!items.length" description="暂无剪切板内容" size="small" />

    <n-scrollbar v-else class="clip-scroll">
      <div v-for="item in items" :key="item.id" class="clip-item">
        <div class="clip-meta">
          <n-tag size="tiny" :type="item.kind === 'image' ? 'warning' : 'info'" :bordered="false">
            {{ item.kind === 'image' ? '图片' : '文本' }}
          </n-tag>
          <n-text depth="3" class="clip-time">{{ fmtTime(item.timestamp) }}</n-text>
          <div class="clip-actions">
            <n-button
              size="tiny"
              quaternary
              :title="isShown(item.id) ? '隐藏内容' : '显示内容'"
              @click="toggleOne(item.id)"
            >
              <template #icon>
                <n-icon>
                  <EyeOffOutline v-if="isShown(item.id)" />
                  <EyeOutline v-else />
                </n-icon>
              </template>
            </n-button>
            <n-button size="tiny" quaternary title="复制到剪贴板" @click="copy(item)">
              <template #icon>
                <n-icon><CopyOutline /></n-icon>
              </template>
            </n-button>
          </div>
        </div>

        <NEllipsis class="clip-caption" :tooltip="true">{{ item.caption }}</NEllipsis>

        <div v-if="!isShown(item.id)" class="clip-mask" @click="toggleOne(item.id)">
          <n-icon><LockClosedOutline /></n-icon>
          <span>{{ item.kind === 'image' && item.preview ? item.preview : '内容已隐藏' }}</span>
        </div>

        <template v-else>
          <div v-if="item.kind === 'text'" class="clip-text">
            {{ texts[item.id] ?? '加载中…' }}
          </div>
          <img v-else-if="images[item.id]" :src="images[item.id]" class="clip-img" alt="" />
          <div v-else class="clip-mask"><n-icon><LockClosedOutline /></n-icon><span>加载中…</span></div>
        </template>
      </div>
    </n-scrollbar>
  </n-card>
</template>

<style scoped>
.clip-panel {
  height: 100%;
}

.clip-panel :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.clip-scroll {
  flex: 1;
  min-height: 0;
}

.clip-item {
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(128, 128, 128, 0.08);
}

.clip-item + .clip-item {
  margin-top: 10px;
}

.clip-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.clip-time {
  font-size: 12px;
}

.clip-actions {
  margin-left: auto;
  display: flex;
  gap: 2px;
}

.clip-caption {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.75;
}

.clip-mask {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px;
  border: 1px dashed rgba(128, 128, 128, 0.45);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
}

.clip-text {
  margin-top: 6px;
  padding: 8px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.04);
  font-family: ui-monospace, SFMono-Regular, Consolas, 'Courier New', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  overflow: auto;
  max-height: 140px;
}

.clip-img {
  margin-top: 6px;
  display: block;
  border-radius: 6px;
  object-fit: contain;
  max-width: 100%;
  max-height: 140px;
}
</style>
