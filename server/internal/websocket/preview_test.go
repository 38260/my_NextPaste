package websocket

import (
	"strings"
	"testing"
)

func TestBuildPreviewText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"普通文本原样", "hello", "hello"},
		{"换行替换为符号", "第一行\r\n第二行\n第三行", "第一行⏎第二行⏎第三行"},
		{"首尾空白去除", "  hello  ", "hello"},
		{"短文本不截断", strings.Repeat("中", 100), strings.Repeat("中", 100)},
		{"超长按 rune 截断", strings.Repeat("中", 200), strings.Repeat("中", 120) + "…"},
	}
	for _, c := range cases {
		if got := buildTextPreview(c.in); got != c.want {
			t.Errorf("%s: buildTextPreview(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestBuildPreviewImage(t *testing.T) {
	data := make([]byte, 1024*1024) // 1.00 MB
	got := buildPreview("image", data)
	if got != "图片 · 1.00 MB" {
		t.Errorf("buildPreview(image) = %q, want %q", got, "图片 · 1.00 MB")
	}
}

func TestBuildPreviewUnknownType(t *testing.T) {
	if got := buildPreview("file", []byte("x")); got != "" {
		t.Errorf("未知类型应返回空预览, got %q", got)
	}
}
