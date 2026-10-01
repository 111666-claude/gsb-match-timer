// Package matchtimer 是对局计时：按真实时间推进、暂停不计时、超时含边界、开始幂等。
// 缺陷：按 Tick 次数扣时间（暂停也扣）、超时用严格小于、重复 Start 会重置计时。
package matchtimer

// MatchLimitMs 是对局时长上限。
const MatchLimitMs = 600000

// Timer 是一局比赛的计时器。
type Timer struct {
	limitMs   int64
	remaining int64
	startedMs int64
	pausedMs  int64
	pauseAtMs int64
	paused    bool
	started   bool
}

// NewTimer 建计时器。
func NewTimer() *Timer { return &Timer{limitMs: MatchLimitMs, remaining: MatchLimitMs} }

// Start 开始计时。缺陷：重复开始会重置进度，而且总是返回 true。
func (t *Timer) Start(nowMs int64) bool {
	t.started = true
	t.startedMs = nowMs
	t.remaining = t.limitMs
	return true
}

// Pause 暂停计时。
func (t *Timer) Pause(nowMs int64) {
	t.paused = true
	t.pauseAtMs = nowMs
}

// Resume 恢复计时。缺陷：暂停的时长没有被累计。
func (t *Timer) Resume(nowMs int64) {
	t.paused = false
}

// Tick 推进一帧。缺陷：不管暂停与否都扣时间，而且只看调用次数。
func (t *Timer) Tick(deltaMs int64) {
	t.remaining -= deltaMs
}

// Remaining 是剩余时间。缺陷：用的是 Tick 累计出来的值，与 nowMs 无关。
func (t *Timer) Remaining(nowMs int64) int64 { return t.remaining }

// Expired 是否超时。缺陷：用严格小于判定，剩余正好为 0 时不算超时。
func (t *Timer) Expired(nowMs int64) bool { return t.remaining < 0 }
