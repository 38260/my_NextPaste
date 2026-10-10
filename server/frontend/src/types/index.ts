export interface LogEntry {
  level: string
  message: string
  preview?: string
  category?: string
  contentId?: number
  timestamp: number
}

export interface ServerStatus {
  isRunning: boolean
  clientCount: number
}

export interface ServerConfig {
  address: string
  port: number
}

// 第三列剪切板面板的条目（由 clipboard 类日志派生）
export interface ClipItem {
  id: number
  kind: 'text' | 'image'
  caption: string
  preview: string
  timestamp: number
}

