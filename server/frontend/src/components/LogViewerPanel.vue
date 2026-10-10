<script setup lang="ts">
import { computed, h, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NEllipsis,
  NSelect,
  NSpace,
  NTag,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { LogEntry } from '../types'

interface Props {
  logs: LogEntry[]
  compact?: boolean
}

interface Emits {
  (e: 'clear'): void
}

const props = withDefaults(defineProps<Props>(), {
  compact: false
})
const emit = defineEmits<Emits>()

// 日志表列：级别 / 分类 / 剪贴板内容 / 时间；原文与图片详情仍在第三列剪切板面板
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
  return new Date(timestamp).toLocaleTimeString('zh-CN', { hour12: false })
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

const columns = computed<DataTableColumns<LogEntry>>(() => [
  {
    title: '级别',
    key: 'level',
    width: props.compact ? 88 : 100,
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
    width: props.compact ? 76 : 84,
    render: (row) => {
      return h(
        NTag,
        { size: 'small', type: categoryTagType(row.category) as any, secondary: true },
        { default: () => categoryLabel(row.category) }
      )
    }
  },
  {
    title: '剪贴板内容',
    key: 'preview',
    render: (row) => {
      const preview = (row.preview ?? '').trim()
      if (preview) {
        return h(NEllipsis, { tooltip: true }, { default: () => preview })
      }
      // 非剪贴板记录没有内容预览，回退为事件描述，避免连接/系统日志整列空白
      return h(NText, { depth: 3 }, { default: () => row.message })
    }
  },
  {
    title: '时间',
    key: 'timestamp',
    width: props.compact ? 96 : 120,
    render: (row) => formatTime(row.timestamp)
  }
])

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
