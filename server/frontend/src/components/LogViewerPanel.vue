<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NEllipsis,
  NInputNumber,
  NPopover,
  NSelect,
  NSpace,
  NSlider,
  NSwitch,
  NTag,
  NText
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { LogEntry } from '../types'
import ClipboardContentCell from './ClipboardContentCell.vue'
import { DEFAULT_SIZES, SIZE_LIMITS, clampSize, loadSizes, saveSizes, type ClipSizes } from '../utils/clipLayout'

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

const sizes = ref<ClipSizes>(loadSizes())
watch(sizes, (value) => saveSizes(value), { deep: true })
const sizeRows = (Object.keys(DEFAULT_SIZES) as (keyof ClipSizes)[]).map((key) => ({
  key,
  label: { textW: '文本框宽', textH: '文本框高', imgW: '图片宽', imgH: '图片高' }[key],
  min: SIZE_LIMITS[key][0],
  max: SIZE_LIMITS[key][1]
}))
const setSize = (key: keyof ClipSizes, value: number | null) => {
  sizes.value = { ...sizes.value, [key]: clampSize(key, Number(value)) }
}
const resetSizes = () => { sizes.value = { ...DEFAULT_SIZES } }

const MASK_KEY = 'np:clipReveal'
const revealAll = ref(localStorage.getItem(MASK_KEY) === '1')
const overrides = ref<Record<number, boolean>>({})
const isShown = (id: number) => overrides.value[id] ?? revealAll.value
const toggleAll = (value: boolean) => {
  revealAll.value = value
  overrides.value = {}
  localStorage.setItem(MASK_KEY, value ? '1' : '0')
}
const toggleOne = (id: number) => { overrides.value[id] = !isShown(id) }
watch(() => props.logs, (logs) => {
  const retained = new Set(logs.map((row) => row.contentId))
  for (const id of Object.keys(overrides.value)) {
    if (!retained.has(Number(id))) delete overrides.value[Number(id)]
  }
})

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
    key: 'content',
    render: (row) => {
      if (row.category === 'clipboard' && row.contentId) {
        return h(ClipboardContentCell, {
          key: row.contentId,
          row,
          sizes: sizes.value,
          shown: isShown(row.contentId),
          onToggle: () => toggleOne(row.contentId!)
        })
      }
      const content = row.category === 'clipboard' && row.preview
        ? revealAll.value ? row.preview : '内容已隐藏'
        : row.message
      return h(NEllipsis, {
        tooltip: {
          contentStyle: {
            maxWidth: 'min(480px, calc(100vw - 32px))',
            maxHeight: 'min(320px, calc(100vh - 32px))',
            overflow: 'auto',
            whiteSpace: 'pre-wrap',
            overflowWrap: 'anywhere'
          }
        }
      }, { default: () => content })
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
  <n-card size="small" title="应用日志" class="log-panel">
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

    <div class="content-controls">
      <n-text depth="3">显示内容</n-text>
      <n-switch size="small" :value="revealAll" aria-label="显示全部剪贴板内容" @update:value="toggleAll" />
      <n-popover trigger="click" placement="bottom-start" :show-arrow="false" style="width: 292px">
        <template #trigger><n-button size="tiny" quaternary title="自定义条目尺寸">尺寸</n-button></template>
        <div class="size-form">
          <div v-for="row in sizeRows" :key="row.key" class="size-row">
            <n-text depth="3" class="size-label">{{ row.label }}</n-text>
            <n-slider :value="sizes[row.key]" :min="row.min" :max="row.max" :step="2" @update:value="(value: number) => setSize(row.key, value)" />
            <n-input-number size="tiny" :value="sizes[row.key]" :min="row.min" :max="row.max" :show-button="false" @update:value="(value: number | null) => setSize(row.key, value)" />
          </div>
          <n-button size="tiny" block secondary @click="resetSizes">恢复默认尺寸</n-button>
        </div>
      </n-popover>
    </div>

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

<style scoped>
.content-controls { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.size-form { display: flex; flex-direction: column; gap: 10px; padding: 4px 2px; }
.size-row { display: flex; align-items: center; gap: 10px; }
.size-label { flex: 0 0 56px; font-size: 12px; }
.size-row :deep(.n-slider) { flex: 1; min-width: 0; }
.size-row :deep(.n-input-number) { flex: 0 0 62px; }
.log-panel :deep(.n-card-header) { flex-wrap: wrap; gap: 8px; }
</style>
