package main

import (
	"fmt"
	"testing"
	"time"
)

// fakeClock 可推进的假时钟，避免测试依赖真实等待
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestLedger() (*clipLedger, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)}
	l := newClipLedger()
	l.now = clock.now
	return l, clock
}

// TestNoteLoggedDedupsSameContent 一条同步链路里多层各记一行（收到 / 已写入 / 广播）时，
// 只保留最先到达的那一行
func TestNoteLoggedDedupsSameContent(t *testing.T) {
	l, _ := newTestLedger()

	content := []byte("同一份文本")
	sig := clipSig("text", content)

	if !l.noteLogged(sig) {
		t.Fatal("首次出现应入日志")
	}
	if l.noteLogged(sig) {
		t.Fatal("窗口内同一内容应被去重")
	}
}

// TestNoteLoggedAllowsAfterTTL 窗口结束后用户真的再复制一次同样内容，仍应出现新条目
func TestNoteLoggedAllowsAfterTTL(t *testing.T) {
	l, clock := newTestLedger()
	sig := clipSig("text", []byte("ok"))

	if !l.noteLogged(sig) {
		t.Fatal("首次出现应入日志")
	}
	clock.advance(clipLedgerTTL + time.Second)
	if !l.noteLogged(sig) {
		t.Fatal("超出窗口后同内容应重新入日志")
	}
}

// TestClipSigSeparatesTypeAndBytes 同字节数不同内容、同内容不同类型都必须是不同签名
func TestClipSigSeparatesTypeAndBytes(t *testing.T) {
	if clipSig("text", []byte("a")) == clipSig("text", []byte("b")) {
		t.Fatal("不同内容签名相同")
	}
	if clipSig("text", []byte("a")) == clipSig("image", []byte("a")) {
		t.Fatal("同字节不同类别签名相同")
	}
	if clipSig("text", []byte("a")) != clipSig("text", []byte("a")) {
		t.Fatal("同内容签名不稳定")
	}
}

// TestSelfEchoSuppression 本端写回剪贴板触发的监听上报应识别为回声；
// 用户自己复制的内容不受影响
func TestSelfEchoSuppression(t *testing.T) {
	l, clock := newTestLedger()

	remote := clipSig("text", []byte("来自手机"))
	l.markWritten(remote)
	if !l.isSelfEcho(remote) {
		t.Fatal("刚写回的内容应判为回声")
	}

	other := clipSig("text", []byte("用户新复制"))
	if l.isSelfEcho(other) {
		t.Fatal("未写回过的内容不应判为回声")
	}

	clock.advance(clipLedgerTTL + time.Second)
	if l.isSelfEcho(remote) {
		t.Fatal("超出窗口后不应继续判为回声")
	}
}

// TestLedgerBounded 台账按容量淘汰，长时间运行不会无限增长
func TestLedgerBounded(t *testing.T) {
	l, _ := newTestLedger()

	for i := 0; i < clipLedgerMax+10; i++ {
		l.noteLogged(clipSig("text", []byte(fmt.Sprintf("item-%d", i))))
	}
	if len(l.logged) > clipLedgerMax {
		t.Fatalf("logged 未淘汰: %d > %d", len(l.logged), clipLedgerMax)
	}
	if len(l.loggedOrder) > clipLedgerMax {
		t.Fatalf("loggedOrder 未淘汰: %d > %d", len(l.loggedOrder), clipLedgerMax)
	}
}
