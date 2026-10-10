<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { NButton, NIcon, NImage, useMessage } from 'naive-ui'
import { CopyOutline, EyeOffOutline, EyeOutline } from '@vicons/ionicons5'
import { CopyLogContent, GetLogContent } from '../../wailsjs/go/main/App'
import type { LogEntry } from '../types'
import type { ClipSizes } from '../utils/clipLayout'

const props = defineProps<{ row: LogEntry; sizes: ClipSizes; shown: boolean }>()
const emit = defineEmits<{ (e: 'toggle'): void }>()
const message = useMessage()
const text = ref<string>()
const imageUrl = ref('')
const bytes = ref(0)
const dimensions = ref('')
const error = ref('')
let loaded = false
let pending = false
let active = true
onBeforeUnmount(() => { active = false })

const isImage = computed(() => Boolean(imageUrl.value) || props.row.preview?.startsWith('图片'))
const contentStyle = computed(() => isImage.value
  ? { width: `${props.sizes.imgW}px`, height: `${props.sizes.imgH}px` }
  : { width: `${props.sizes.textW}px`, maxHeight: `${props.sizes.textH}px` })

// 单元格按内容 ID 设置 key；隐藏态不读取正文，卸载后丢弃未完成的读取。
watch(() => props.shown, async (shown) => {
  if (!shown || loaded || pending) return
  pending = true
  error.value = ''
  try {
    const result = await GetLogContent(props.row.contentId!)
    if (!active) return
    if (!result?.found) {
      error.value = result?.error || '内容不可用'
    } else if (result.type === 'text') {
      text.value = result.text ?? ''
      loaded = true
    } else if (result.type === 'image' && result.imageBase64) {
      imageUrl.value = `data:image/png;base64,${result.imageBase64}`
      bytes.value = Number(result.size) || 0
      loaded = true
    } else {
      error.value = '内容不可用'
    }
  } catch {
    if (active) error.value = '读取内容失败，请重试'
  } finally {
    pending = false
  }
}, { immediate: true })

const copy = async () => {
  try {
    await CopyLogContent(props.row.contentId!)
    message.success('已复制到剪贴板')
  } catch (error) {
    message.error(`复制失败：${String(error)}`)
  }
}

const measure = (event: Event) => {
  const image = event.target as HTMLImageElement
  dimensions.value = `${image.naturalWidth} × ${image.naturalHeight} px`
}

const formatBytes = (value: number) => value < 1024
  ? `${value} B`
  : value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(2)} MB`
</script>

<template>
  <div class="clip-cell">
    <div class="clip-body">
      <n-button
        v-if="!shown"
        class="clip-mask"
        :style="contentStyle"
        aria-label="显示内容"
        @click="emit('toggle')"
      >{{ isImage ? row.preview : '内容已隐藏' }}</n-button>
      <div v-else-if="error" class="clip-mask" :style="contentStyle" role="alert">{{ error }}</div>
      <template v-else-if="imageUrl">
        <n-image
          :src="imageUrl"
          :width="sizes.imgW"
          :height="sizes.imgH"
          object-fit="contain"
          :img-props="{ style: { maxWidth: '100%' }, alt: '剪贴板图片', onLoad: measure }"
          class="clip-image"
        />
        <div class="clip-dim">{{ dimensions }} · {{ formatBytes(bytes) }} · 点击可放大</div>
      </template>
      <div v-else class="clip-text" :style="contentStyle">{{ text ?? '加载中…' }}</div>
    </div>
    <div class="clip-actions">
      <n-button size="tiny" quaternary :title="shown ? '隐藏内容' : '显示内容'" :aria-label="shown ? '隐藏内容' : '显示内容'" @click="emit('toggle')">
        <template #icon><n-icon><EyeOffOutline v-if="shown" /><EyeOutline v-else /></n-icon></template>
      </n-button>
      <n-button size="tiny" quaternary title="复制到剪贴板" aria-label="复制到剪贴板" @click="copy">
        <template #icon><n-icon><CopyOutline /></n-icon></template>
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.clip-cell { display: flex; align-items: flex-start; gap: 8px; min-width: 0; }
.clip-body { min-width: 0; }
.clip-actions { display: flex; flex-shrink: 0; gap: 2px; }
.clip-mask { max-width: 100%; min-height: 40px; padding: 8px; border: 1px dashed rgba(128, 128, 128, 0.45); border-radius: 6px; box-sizing: border-box; }
.clip-text { padding: 8px; max-width: 100%; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; background: rgba(128, 128, 128, 0.08); border-radius: 6px; box-sizing: border-box; font-family: ui-monospace, SFMono-Regular, Consolas, 'Courier New', monospace; font-size: 12px; }
.clip-image { max-width: 100%; }
.clip-image :deep(img) { max-width: 100%; border-radius: 6px; cursor: zoom-in; }
.clip-dim { margin-top: 4px; font-size: 11px; opacity: 0.65; overflow-wrap: anywhere; }
</style>
