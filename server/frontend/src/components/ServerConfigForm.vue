<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NButton, NCard, NForm, NFormItem, NInput, NInputNumber } from 'naive-ui'
import type { ServerConfig } from '../types'

interface Props {
  config: ServerConfig
  isRunning: boolean
}

interface Emits {
  (e: 'start', config: ServerConfig): void
  (e: 'stop'): void
  (e: 'update:config', config: ServerConfig): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const localConfig = ref<ServerConfig>({ ...props.config })

watch(
  () => props.config,
  (newConfig) => {
    localConfig.value = { ...newConfig }
  },
  { deep: true }
)

watch(
  localConfig,
  (newConfig) => {
    emit('update:config', newConfig)
  },
  { deep: true }
)

const isValid = computed(() => {
  return (
    localConfig.value.address.length > 0 &&
    localConfig.value.port > 0 &&
    localConfig.value.port <= 65535
  )
})

const handleStart = () => {
  if (!isValid.value) return
  emit('start', localConfig.value)
}

const handleStop = () => {
  emit('stop')
}
</script>

<template>
  <n-card size="small" title="服务器配置">
    <div class="config-row">
      <n-form class="config-fields" label-placement="top" :show-feedback="false" :disabled="isRunning">
        <n-form-item label="监听地址">
          <n-input v-model:value="localConfig.address" placeholder="0.0.0.0" />
        </n-form-item>

        <n-form-item label="端口号">
          <n-input-number v-model:value="localConfig.port" :min="1" :max="65535" style="width: 100%" />
        </n-form-item>
      </n-form>

      <n-button
        v-if="!isRunning"
        type="primary"
        class="config-action"
        :disabled="!isValid"
        @click="handleStart"
      >
        启动服务
      </n-button>
      <n-button v-else type="error" class="config-action" @click="handleStop">停止服务</n-button>
    </div>
  </n-card>
</template>

<style scoped>
.config-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 12px;
}

.config-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 120px;
  gap: 12px;
}

.config-action {
  align-self: end;
}

@media (max-width: 600px) {
  .config-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .config-fields {
    grid-template-columns: minmax(0, 1fr) 110px;
  }
}
</style>
