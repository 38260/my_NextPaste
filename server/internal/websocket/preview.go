package websocket

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// 预览最大字符数（按 rune 计，UTF-8 安全）
const previewMaxRunes = 120

// buildTextPreview 生成文本预览：换行替换为可见符号，超过 120 字符按 rune 截断
func buildTextPreview(text string) string {
	t := strings.ReplaceAll(text, "\r\n", "\n")
	t = strings.ReplaceAll(t, "\n", "⏎")
	t = strings.ReplaceAll(t, "\t", " ")
	t = strings.TrimSpace(t)
	if utf8.RuneCountInString(t) <= previewMaxRunes {
		return t
	}
	runes := []rune(t)
	return string(runes[:previewMaxRunes]) + "…"
}

// BuildPreview 根据剪贴板数据类型生成日志预览列内容（供 websocket 与 app 层共用）
func BuildPreview(dataType string, data []byte) string {
	switch dataType {
	case "text":
		return buildTextPreview(string(data))
	case "image":
		return fmt.Sprintf("图片 · %.2f MB", float64(len(data))/1024/1024)
	default:
		return ""
	}
}
