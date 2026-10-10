// 剪贴板单元格尺寸配置：文本框与图片的宽高由用户自定义并持久化

export interface ClipSizes {
  textW: number
  textH: number
  imgW: number
  imgH: number
}

export const DEFAULT_SIZES: ClipSizes = { textW: 320, textH: 140, imgW: 200, imgH: 140 }

export const SIZE_LIMITS: Record<keyof ClipSizes, [number, number]> = {
  textW: [120, 720],
  textH: [40, 480],
  imgW: [60, 720],
  imgH: [40, 480]
}

const SIZES_KEY = 'np:clipSizes'

const clamp = (v: number, [min, max]: [number, number]) => {
  if (!Number.isFinite(v)) return min
  return Math.min(max, Math.max(min, Math.round(v)))
}

export function loadSizes(): ClipSizes {
  try {
    const raw = localStorage.getItem(SIZES_KEY)
    if (!raw) return { ...DEFAULT_SIZES }
    const obj = JSON.parse(raw) as Partial<ClipSizes>
    return (Object.keys(DEFAULT_SIZES) as (keyof ClipSizes)[]).reduce((acc, key) => {
      acc[key] = clamp(Number(obj?.[key]), SIZE_LIMITS[key])
      return acc
    }, { ...DEFAULT_SIZES })
  } catch {
    return { ...DEFAULT_SIZES }
  }
}

export function saveSizes(sizes: ClipSizes) {
  localStorage.setItem(SIZES_KEY, JSON.stringify(sizes))
}

export function clampSize(key: keyof ClipSizes, value: number) {
  return clamp(Number(value), SIZE_LIMITS[key])
}
