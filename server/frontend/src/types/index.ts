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

