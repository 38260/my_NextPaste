export interface LogEntry {
  level: string
  message: string
  preview?: string
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

