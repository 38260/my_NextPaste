<script setup lang="ts">
import { computed } from 'vue'
import {
  NButton,
  NCheckbox,
  NDivider,
  NForm,
  NFormItem,
  NModal,
  NRadio,
  NRadioGroup,
  NSpace,
  NText,
} from 'naive-ui'
import { Quit } from '../../wailsjs/go/main/App'

// ... 类型定义保持不变 ...
type AutoStartMode = 'server' | 'client'
type Settings = {
  autoStartEnabled: boolean
  autoStartMode: AutoStartMode
  autoStartActionEnabled: boolean
  closeAction: 'minimize' | 'close'
}

interface Props {
  show: boolean
  settings: Settings
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'update:settings', value: Settings): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const show = computed({
  get: () => props.show,
  set: (v: boolean) => emit('update:show', v)
})

const settings = computed({
  get: () => props.settings,
  set: (v: Settings) => emit('update:settings', v)
})

const handleQuitApp = async () => {
  try {
    await Quit()
  } catch {
    // ignore
  }
}
</script>

<template>
  <n-modal 
    v-model:show="show" 
    preset="card" 
    title="系统设置" 
    style="width: 480px"
    :segmented="{ content: 'soft' }"
    size="medium"
  >
    <n-form
      label-placement="left"
      :label-width="120"
      label-align="left"
      :show-feedback="false"
      size="small"
    >
      <div class="settings-section">
        <n-text depth="3" strong class="section-title">启动行为</n-text>
        
        <n-form-item label="开机启动">
          <div>
            <n-checkbox v-model:checked="settings.autoStartEnabled">
              随操作系统自动启动
            </n-checkbox>
            <div style="margin-top: 4px; font-size: 12px">
              <n-text depth="3">开机后静默最小化启动，建议同时开启下方"启动后自动执行"</n-text>
            </div>
          </div>
        </n-form-item>

        <n-form-item label="启动后自动执行">
          <n-checkbox v-model:checked="settings.autoStartActionEnabled">
            自动创建服务端 / 连接客户端
          </n-checkbox>
        </n-form-item>

        <n-form-item
          v-if="settings.autoStartActionEnabled"
          label="执行模式"
          style="margin-top: -8px"
        >
          <n-radio-group v-model:value="settings.autoStartMode" name="mode">
            <n-radio value="server">服务端</n-radio>
            <n-radio value="client">客户端</n-radio>
          </n-radio-group>
        </n-form-item>
      </div>

      <n-divider style="margin: 16px 0" />

      <div class="settings-section">
        <n-text depth="3" strong class="section-title">窗口行为</n-text>

        <n-form-item label="关闭按钮">
          <n-radio-group v-model:value="settings.closeAction" name="action">
            <n-space vertical :size="4">
              <n-radio value="minimize">最小化到任务栏（后台保持同步）</n-radio>
              <n-radio value="close">退出应用程序</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <n-form-item label="说明">
          <n-text depth="3" style="font-size: 12px">
            最小化驻留时，任务栏图标可随时点回窗口；需要完全退出请点右下角"退出应用"
          </n-text>
        </n-form-item>
      </div>
    </n-form>

    <template #footer>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <n-text depth="3" style="font-size: 12px">
          修改将实时保存并生效
        </n-text>
        <n-button size="small" type="error" secondary @click="handleQuitApp">
          退出应用
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.section-title {
  display: block;
  font-size: 12px;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.settings-section {
  padding: 4px 8px;
}

:deep(.n-form-item) {
  margin-bottom: 12px;
}

:deep(.n-form-item:last-child) {
  margin-bottom: 0;
}
</style>