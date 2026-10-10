package main

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// clipLedgerTTL 同一内容签名在该时间窗内只入一条日志、只认一次回声
const clipLedgerTTL = 3 * time.Second

// clipLedgerMax 台账容量，防止长跑内存增长
const clipLedgerMax = 64

// clipSig 内容签名：类型 + 原始字节摘要（同内容不同文本表示同签名）
func clipSig(dataType string, content []byte) string {
	h := md5.Sum(content)
	return fmt.Sprintf("%s:%x", dataType, h)
}

// clipLedger 剪贴板内容台账，解决两类重复：
//  1. 同一条内容在一次同步链路里被多层各记一行（ws 层"收到…" + App 层"已接收并写入…"，
//     或本地"检测到…" + "广播…"）；
//  2. 远程内容写回系统剪贴板后，本地监听把它当作用户新复制再次广播（回声，
//     会在两端之间来回出现，且 monitor 的类型内 hash 守卫挡不住跨类型触发）。
type clipLedger struct {
	mu sync.Mutex
	// logged 内容签名 -> 最近一次入日志时间
	logged map[string]time.Time
	// loggedOrder 按写入顺序淘汰
	loggedOrder []string
	// written 由本端主动写回剪贴板的内容签名（回声源）
	written map[string]time.Time
	// now 可注入以便测试
	now func() time.Time
}

func newClipLedger() *clipLedger {
	return &clipLedger{
		logged:  make(map[string]time.Time),
		written: make(map[string]time.Time),
		now:     time.Now,
	}
}

// noteLogged 登记一条剪贴板日志；返回 false 表示窗口内已出现过同一内容，调用方应丢弃。
func (l *clipLedger) noteLogged(sig string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if at, ok := l.logged[sig]; ok && now.Sub(at) < clipLedgerTTL {
		return false
	}
	if len(l.loggedOrder) >= clipLedgerMax {
		oldest := l.loggedOrder[0]
		l.loggedOrder = l.loggedOrder[1:]
		delete(l.logged, oldest)
	}
	l.logged[sig] = now
	l.loggedOrder = append(l.loggedOrder, sig)
	return true
}

// markWritten 标记"这份内容是我自己写回系统剪贴板的"，窗口内监听再上报即视为回声。
func (l *clipLedger) markWritten(sig string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.written[sig] = l.now()
	if len(l.written) > clipLedgerMax {
		for k, at := range l.written {
			if l.now().Sub(at) >= clipLedgerTTL {
				delete(l.written, k)
			}
		}
	}
}

// isSelfEcho 判断监听上报的内容是否来自本端刚写回的剪贴板。
func (l *clipLedger) isSelfEcho(sig string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	at, ok := l.written[sig]
	return ok && l.now().Sub(at) < clipLedgerTTL
}
