// 客户端连接配置模板：localStorage 持久化，内置默认模板固定存在
export interface ConnTemplate {
  id: string
  name: string
  url: string
  builtin: boolean
}

const KEY = 'np:connTemplates'

// 内置默认模板（builtin 固定存在、不可删除）
const BUILTIN_TEMPLATES: ConnTemplate[] = [
  { id: 'builtin-local', name: '本机测试', url: 'ws://localhost:8080/ws/my-room', builtin: true },
  { id: 'builtin-lan', name: '局域网服务器', url: 'ws://192.168.1.100:8080/ws/my-room', builtin: true }
]

export function loadConnTemplates(): ConnTemplate[] {
  try {
    const raw = localStorage.getItem(KEY)
    if (!raw) {
      localStorage.setItem(KEY, JSON.stringify(BUILTIN_TEMPLATES))
      return [...BUILTIN_TEMPLATES]
    }

    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) {
      return [...BUILTIN_TEMPLATES]
    }

    // 只保留合法条目，并把缺失的内置模板补回去（固定默认记忆）
    const list: ConnTemplate[] = []
    for (const item of parsed) {
      if (
        item &&
        typeof item.id === 'string' &&
        typeof item.name === 'string' &&
        typeof item.url === 'string'
      ) {
        list.push({
          id: item.id,
          name: item.name,
          url: item.url,
          builtin: Boolean(item.builtin)
        })
      }
    }
    for (const builtin of BUILTIN_TEMPLATES) {
      if (!list.some((t) => t.id === builtin.id)) {
        list.unshift(builtin)
      }
    }
    return list
  } catch {
    return [...BUILTIN_TEMPLATES]
  }
}

export function saveConnTemplates(list: ConnTemplate[]) {
  localStorage.setItem(KEY, JSON.stringify(list))
}

// 从 ws 地址提取默认模板名（host:port）
export function templateNameFromUrl(url: string): string {
  try {
    return new URL(url).host
  } catch {
    return url
  }
}
