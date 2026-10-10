<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NSpace,
  NTag,
  NText,
  NTooltip,
  useMessage
} from 'naive-ui'
import {
  loadConnTemplates,
  saveConnTemplates,
  templateNameFromUrl,
  type ConnTemplate
} from '../utils/connTemplates'

interface Props {
  url: string
  isConnected: boolean
}

interface Emits {
  (e: 'connect', url: string): void
  (e: 'disconnect'): void
  (e: 'update:url', url: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const message = useMessage()

const localUrl = ref(props.url)

watch(
  () => props.url,
  (newUrl) => {
    localUrl.value = newUrl
  }
)

watch(localUrl, (newUrl) => {
  emit('update:url', newUrl)
})

const isValid = computed(() => {
  return localUrl.value.startsWith('ws://') || localUrl.value.startsWith('wss://')
})

const handleConnect = () => {
  if (!isValid.value) return
  emit('connect', localUrl.value)
}

const handleDisconnect = () => {
  emit('disconnect')
}

// 连接模板
const templates = ref<ConnTemplate[]>(loadConnTemplates())

const saving = ref(false)
const saveName = ref('')

const handleApplyTemplate = (tpl: ConnTemplate) => {
  if (props.isConnected) {
    message.warning('已连接状态下请先断开，再切换模板')
    return
  }
  localUrl.value = tpl.url
  emit('update:url', tpl.url)
  emit('connect', tpl.url)
}

const handleStartSave = () => {
  if (!isValid.value) {
    message.warning('请先输入有效的 ws:// 地址')
    return
  }
  saveName.value = templateNameFromUrl(localUrl.value)
  saving.value = true
}

const handleConfirmSave = () => {
  const name = saveName.value.trim()
  if (!name) {
    message.warning('模板名称不能为空')
    return
  }
  templates.value.push({
    id: `tpl-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name,
    url: localUrl.value,
    builtin: false
  })
  saveConnTemplates(templates.value)
  saving.value = false
  message.success(`已保存模板「${name}」`)
}

const handleCancelSave = () => {
  saving.value = false
}

const handleRemoveTemplate = (tpl: ConnTemplate) => {
  if (tpl.builtin) return
  templates.value = templates.value.filter((t) => t.id !== tpl.id)
  saveConnTemplates(templates.value)
  message.info(`已删除模板「${tpl.name}」`)
}
</script>

<template>
  <n-card size="small" title="客户端配置">
    <n-form label-placement="top" :disabled="isConnected">
      <n-form-item label="服务器地址">
        <n-input v-model:value="localUrl" placeholder="ws://server:8080/ws" />
      </n-form-item>
    </n-form>

    <n-button
      v-if="!isConnected"
      type="primary"
      block
      :disabled="!isValid"
      @click="handleConnect"
    >
      连接服务器
    </n-button>
    <n-button v-else type="error" block @click="handleDisconnect">断开连接</n-button>

    <n-space v-if="templates.length" vertical :size="6" style="margin-top: 16px">
      <n-text depth="3" style="font-size: 12px">
        连接模板：点击即填入并连接（内置模板固定保留）
      </n-text>

      <n-space :size="6">
        <n-tooltip v-for="tpl in templates" :key="tpl.id" trigger="hover">
          <template #trigger>
            <n-tag
              size="small"
              :closable="!tpl.builtin"
              :bordered="!tpl.builtin"
              :type="tpl.builtin ? 'default' : 'info'"
              style="cursor: pointer"
              @click="handleApplyTemplate(tpl)"
              @close="handleRemoveTemplate(tpl)"
            >
              {{ tpl.name }}
            </n-tag>
          </template>
          {{ tpl.url }}
        </n-tooltip>
      </n-space>

      <template v-if="!isConnected">
        <n-space v-if="!saving" :size="6" align="center" style="margin-top: 2px">
          <n-button size="tiny" quaternary @click="handleStartSave">把当前地址存为模板</n-button>
        </n-space>
        <n-space v-else :size="6" align="center">
          <n-input
            v-model:value="saveName"
            size="tiny"
            style="width: 160px"
            placeholder="模板名称"
            @keyup.enter="handleConfirmSave"
          />
          <n-button size="tiny" type="primary" @click="handleConfirmSave">保存</n-button>
          <n-button size="tiny" quaternary @click="handleCancelSave">取消</n-button>
        </n-space>
      </template>
    </n-space>
  </n-card>
</template>
