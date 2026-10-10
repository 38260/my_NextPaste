<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NEllipsis,
  NIcon,
  NSelect,
  NSpace,
  NTag,
  NText,
  useMessage
} from 'naive-ui'
import { CopyOutline } from '@vicons/ionicons5'
import { CopyLogContent, GetLogContent } from '../../wailsjs/go/main/App'
import type { DataTableColumns } from 'naive-ui'
import type { LogEntry } from '../types'

interface Props {
  logs: LogEntry[]
}

interface Emits {
  (e: 'clear'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const message = useMessage()

// 图片缩略图懒加载缓存：contentId -> data URL
const imageCache = ref<Record<number, string>>({})
const imageLoading = new Set<number>()

const ensureImageLoaded = (contentId: number) => {
  if (imageCache.value[contentId] || imageLoading.has(contentId)) return
  imageLoading.add(contentId)
  GetLogContent(contentId)
    .then((res: any) => {
      if (res && res.found && res.type === 'image' && res.imageBase64) {
        imageCache.value = {
          ...imageCache.value,
          [contentId]: `data:image/png;base64,${res.imageBase64}`
        }
      }
    })
    .catch(() => {})
    .finally(() => imageLoading.delete(contentId))
}

const copyContent = async (contentId: number) => {
  try {
    await CopyLogContent(contentId)
    message.success('已复制到剪贴板')
  } catch (error) {
    message.error(`复制失败：${String(error)}`)
  }
}

type Level = 'ALL' | 'INFO' | 'SUCCESS' | 'WARNING' | 'ERROR'
type Category = 'ALL' | 'clipboard' | 'connection' | 'system'

const filterLevel = ref<Level>('ALL')
// 默认只看剪贴板相关内容
const filterCategory = ref<Category>('clipboard')

const filteredLogs = computed(() => {
  return props.logs.filter((l) => {
    if (filterCategory.value !== 'ALL' && l.category !== filterCategory.value) return false
    if (filterLevel.value !== 'ALL' && l.level !== filterLevel.value) return false
    return true
  })
})

const formatTime = (timestamp: number) => {
  if (!timestamp) {
    const now = new Date()
    return now.toLocaleTimeString('zh-CN', { hour12: false })
  }
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour12: false })
}

const levelTagType = (level: string) => {
  switch (level) {
    case 'SUCCESS':
      return 'success'
    case 'WARNING':
      return 'warning'
    case 'ERROR':
      return 'error'
    case 'INFO':
      return 'info'
    default:
      return 'default'
  }
}

const categoryLabel = (category?: string) => {
  switch (category) {
    case 'clipboard':
      return '剪贴板'
    case 'connection':
      return '连接'
    case 'system':
      return '系统'
    default:
      return '系统'
  }
}

const categoryTagType = (category?: string) => {
  switch (category) {
    case 'clipboard':
      return 'success'
    case 'connection':
      return 'info'
    default:
      return 'default'
  }
}

const columns: DataTableColumns<LogEntry> = [
  {
    title: '级别',
    key: 'level',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        { size: 'small', type: levelTagType(row.level) as any },
        { default: () => row.level }
      )
    }
  },
  {
    title: '分类',
    key: 'category',
    width: 84,
    render: (row) => {
      return h(
        NTag,
        { size: 'small', type: categoryTagType(row.category) as any, secondary: true },
        { default: () => categoryLabel(row.category) }
      )
    }
  },
  {
    title: '内容',
    key: 'message',
    width: 280,
    render: (row) => {
      return h(NEllipsis, { tooltip: true }, { default: () => row.message })
    }
  },
  {
    title: '剪贴板内容',
    key: 'preview',
    width: 260,
    render: (row) => {
      const parts: any[] = []

      if (row.contentId && row.preview && row.preview.startsWith('图片')) {
        const url = imageCache.value[row.contentId]
        if (url) {
          parts.push(
            h('img', {
              src: url,
              style:
                'max-height:48px; max-width:140px; border-radius:4px; display:block; flex-shrink:0'
            })
          )
        } else {
          ensureImageLoaded(row.contentId)
          parts.push(h('span', { style: 'opacity: 0.5; font-size: 12px' }, '加载中…'))
        }
      } else if (row.preview) {
        parts.push(
          h(
            NEllipsis,
            {
              tooltip: true,
              style:
                'font-family: ui-monospace, SFMono-Regular, Consolas, "Courier New", monospace; font-size: 12px'
            },
            { default: () => row.preview }
          )
        )
      } else {
        parts.push(h('span', { style: 'opacity: 0.35; user-select: none' }, '—'))
      }

      if (row.contentId) {
        parts.push(
          h(
            NButton,
            {
              size: 'tiny',
              quaternary: true,
              style: 'flex-shrink: 0',
              title: '复制此内容',
              onClick: () => copyContent(row.contentId as number)
            },
            { icon: () => h(NIcon, null, { default: () => h(CopyOutline) }) }
          )
        )
      }

      return h('div', { style: 'display: flex; align-items: center; gap: 6px; min-width: 0' }, parts)
    }
  },
  {
    title: '时间',
    key: 'timestamp',
    width: 120,
    render: (row) => formatTime(row.timestamp)
  }
]

// avoid unused import for h by defining here
import { h } from 'vue'

watch(
  () => props.logs.length,
  () => {
    // DataTable itself handles scrolling; keep behavior unchanged (no forced scroll)
  }
)

const handleClear = () => emit('clear')
</script>

<template>
  <n-card size="small" title="应用日志">
    <template #header-extra>
      <n-space align="center" :size="8">
        <n-text depth="3">{{ filteredLogs.length }} 条</n-text>
        <n-select
          v-model:value="filterCategory"
          size="small"
          style="width: 110px"
          :options="[
            { label: '全部分类', value: 'ALL' },
            { label: '剪贴板', value: 'clipboard' },
            { label: '连接', value: 'connection' },
            { label: '系统', value: 'system' }
          ]"
        />
        <n-select
          v-model:value="filterLevel"
          size="small"
          style="width: 120px"
          :options="[
            { label: '全部', value: 'ALL' },
            { label: '信息', value: 'INFO' },
            { label: '成功', value: 'SUCCESS' },
            { label: '警告', value: 'WARNING' },
            { label: '错误', value: 'ERROR' }
          ]"
        />
        <n-button size="small" @click="handleClear">清空</n-button>
      </n-space>
    </template>

    <n-data-table
      size="small"
      :columns="columns"
      :data="filteredLogs"
      :bordered="true"
      :single-line="false"
      :max-height="520"
    />
  </n-card>
</template>
