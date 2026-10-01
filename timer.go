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
// 剩余时间不逐帧扣减，而是在查询时由开始时刻、当前时刻与累计暂停时长即时推导，
// 因此同一串事件重放结果必然一致，每次操作都是 O(1)。
type Timer struct {
	started      bool
	startedMs    int64
	lastMs       int64
	paused       bool
	pauseStartMs int64
	pausedMs     int64
	pauses       int
}

// NewTimer 建时钟。
func NewTimer() *Timer {
	return &Timer{}
}

// Start 开始计时。重复开始返回 false 且不重置任何进度。
func (t *Timer) Start(nowMs int64) bool {
	if t.started {
		return false
	}
	t.started = true
	t.startedMs = nowMs
	t.lastMs = nowMs
	return true
}

// Pause 请求暂停。
// 未开始、已暂停、超过 MaxPauses 配额或已进入加时（含正好归零的边界）时返回 false 且不改状态。
func (t *Timer) Pause(nowMs int64) bool {
	if !t.started || t.paused || t.pauses >= MaxPauses {
		return false
	}
	eff := nowMs
	if eff < t.lastMs {
		eff = t.lastMs
	}
	if t.playedAt(eff) >= MatchLimitMs {
		return false
	}
	t.lastMs = eff
	t.paused = true
	t.pauseStartMs = eff
	t.pauses++
	return true
}

// Resume 恢复计时，把本次暂停时长累计进去；未暂停时是空操作。
func (t *Timer) Resume(nowMs int64) {
	if !t.started || !t.paused {
		return
	}
	eff := nowMs
	if eff < t.lastMs {
		eff = t.lastMs
	}
	t.pausedMs += eff - t.pauseStartMs
	t.paused = false
	t.lastMs = eff
}

// Tick 保留旧入口；时钟按真实时间推进，逐帧回调不再扣减时间。
func (t *Timer) Tick(deltaMs int64) {}

// playedAt 返回时刻 eff 已经消耗的比赛时长（扣除暂停，且不含加时暂停）。
func (t *Timer) playedAt(eff int64) int64 {
	played := eff - t.startedMs - t.pausedMs
	if t.paused {
		played -= eff - t.pauseStartMs
	}
	if played < 0 {
		played = 0
	}
	return played
}

// observe 时钟回拨保护：nowMs 小于上次推进时刻时沿用上一次时刻，并刷新最高水位。
func (t *Timer) observe(nowMs int64) int64 {
	if nowMs > t.lastMs {
		t.lastMs = nowMs
	}
	return t.lastMs
}

// Remaining 是剩余时间：常规时间内返回到常规时间结束的余量，进入加时后返回加时余量（下限为 0）。
func (t *Timer) Remaining(nowMs int64) int64 {
	if !t.started {
		return MatchLimitMs
	}
	played := t.playedAt(t.observe(nowMs))
	if played < MatchLimitMs {
		return MatchLimitMs - played
	}
	remaining := int64(OvertimeLimitMs) - (played - MatchLimitMs)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Overtime 是否处于加时。常规时间正好归零即算进入加时。
func (t *Timer) Overtime(nowMs int64) bool {
	if !t.started {
		return false
	}
	return t.playedAt(t.observe(nowMs)) >= MatchLimitMs
}

// Expired 是否已经打完。加时也走完（含剩余正好为 0 的边界）才算打完。
func (t *Timer) Expired(nowMs int64) bool {
	if !t.started {
		return false
	}
	return t.playedAt(t.observe(nowMs)) >= MatchLimitMs+OvertimeLimitMs
}
