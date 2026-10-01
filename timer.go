// Package matchtimer 是对局时钟：按真实时间推进、暂停有配额、剩余归零进入加时、时钟回拨不推进。
package matchtimer

// 对局限额、加时限额与暂停配额。
const (
	MatchLimitMs    = 600000
	OvertimeLimitMs = 60000
	MaxPauses       = 2
)

// Timer 是一局比赛的时钟。
//
// 剩余时间不保存为可变字段，而是由开始时刻、累计暂停时长与 nowMs 现算，
// 因此同一串事件重放结果一致。所有方法单次 O(1)。
type Timer struct {
	startedMs int64 // 开始时刻
	lastMs    int64 // 上次推进的有效时刻（回拨保护上界）
	pausedMs  int64 // 累计已生效的暂停时长
	pauseAtMs int64 // 当前暂停起点；未暂停为 -1
	pauses    int   // 已接受的暂停次数
	started   bool
	overtime  bool
}

// NewTimer 建时钟。
func NewTimer() *Timer {
	return &Timer{pauseAtMs: -1}
}

// effectiveNow 做时钟回拨保护：nowMs 小于上次推进时刻时不推进。
func (t *Timer) effectiveNow(nowMs int64) int64 {
	if nowMs < t.lastMs {
		return t.lastMs
	}
	return nowMs
}

// advance 把时钟推进到 nowMs：暂停期间不计时，常规时间走完进入加时。
func (t *Timer) advance(nowMs int64) {
	if !t.started {
		return
	}
	now := t.effectiveNow(nowMs)
	if now < t.lastMs {
		return
	}
	if t.pauseAtMs >= 0 {
		if now > t.pauseAtMs {
			t.pausedMs += now - t.pauseAtMs
			t.pauseAtMs = now
		}
	} else if !t.overtime {
		elapsed := now - t.startedMs - t.pausedMs
		if elapsed >= MatchLimitMs {
			t.overtime = true
		}
	}
	t.lastMs = now
}

// Start 开始计时。重复开始返回 false 且不改任何状态。
func (t *Timer) Start(nowMs int64) bool {
	if t.started {
		return false
	}
	t.started = true
	t.startedMs = nowMs
	t.lastMs = nowMs
	return true
}

// Pause 请求暂停。每局最多 MaxPauses 次；加时期间暂停无效；
// 已处于暂停或已打完也不接受。被拒绝时返回 false 且不改状态。
func (t *Timer) Pause(nowMs int64) bool {
	if !t.started || t.overtime || t.paused() || t.pauses >= MaxPauses {
		return false
	}
	now := t.effectiveNow(nowMs)
	elapsed := now - t.startedMs - t.pausedMs
	if elapsed >= MatchLimitMs {
		return false
	}
	t.advance(nowMs)
	if t.overtime {
		return false
	}
	t.pauses++
	t.pauseAtMs = now
	return true
}

// Resume 恢复计时。暂停时长在此刻结算并累计。
func (t *Timer) Resume(nowMs int64) {
	if !t.started || !t.paused() {
		return
	}
	t.advance(nowMs)
	t.pauseAtMs = -1
}

func (t *Timer) paused() bool { return t.pauseAtMs >= 0 }

// remainingAt 返回推进到 nowMs 后的剩余时间（含加时），不在此时推进状态。
func (t *Timer) remainingAt(nowMs int64) int64 {
	if !t.started {
		return MatchLimitMs
	}
	now := t.effectiveNow(nowMs)
	pausedMs := t.pausedMs
	pauseAt := t.pauseAtMs
	if pauseAt >= 0 && now > pauseAt {
		pausedMs += now - pauseAt
		pauseAt = now
	}
	if !t.overtime {
		elapsed := now - t.startedMs - pausedMs
		if elapsed < 0 {
			elapsed = 0
		}
		if elapsed < MatchLimitMs {
			return MatchLimitMs - elapsed
		}
		overElapsed := elapsed - MatchLimitMs
		over := OvertimeLimitMs - overElapsed
		if over < 0 {
			over = 0
		}
		return over
	}
	elapsed := now - t.startedMs - pausedMs - MatchLimitMs
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := OvertimeLimitMs - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Tick 是旧接口的兼容空操作：时间只由真实时间（nowMs）推进，与调用次数无关。
func (t *Timer) Tick(deltaMs int64) {}

// Remaining 是剩余时间（毫秒）；查询本身按真实时间推进时钟。
func (t *Timer) Remaining(nowMs int64) int64 {
	t.advance(nowMs)
	return t.remainingAt(nowMs)
}

// Overtime 是否处于加时；查询本身按真实时间推进时钟。
func (t *Timer) Overtime(nowMs int64) bool {
	t.advance(nowMs)
	return t.overtime
}

// Expired 是否已经打完：常规时间与加时都走完才算打完（含边界）。
func (t *Timer) Expired(nowMs int64) bool {
	t.advance(nowMs)
	return t.remainingAt(nowMs) <= 0
}
