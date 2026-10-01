// Package matchtimer 是对局时钟：按真实时间推进、暂停有配额、剩余归零进入加时、时钟回拨不推进。
// 缺陷：按 Tick 次数扣时间（暂停也扣）、暂停没有配额、剩余归零不进入加时、时钟回拨照扣、重复开始重置。
package matchtimer

// 对局限额、加时限额与暂停配额。
const (
	MatchLimitMs    = 600000
	OvertimeLimitMs = 60000
	MaxPauses       = 2
)

// Timer 是一局比赛的时钟。
type Timer struct {
	remainingMs int64
	ticks       int
	overtime    bool
	pauses      int
	paused      bool
	pausedMs    int64
	pauseAtMs   int64
	startedMs   int64
	started     bool
	lastMs      int64
}

// NewTimer 建时钟。
func NewTimer() *Timer {
	return &Timer{remainingMs: MatchLimitMs}
}

// Start 开始计时。缺陷：重复开始会重置进度，而且总是返回 true。
func (t *Timer) Start(nowMs int64) bool {
	t.started = true
	t.startedMs = nowMs
	t.lastMs = nowMs
	t.remainingMs = MatchLimitMs
	t.overtime = false
	return true
}

// Pause 请求暂停。缺陷：没有配额限制，永远返回 true，也不记暂停起点。
func (t *Timer) Pause(nowMs int64) bool {
	t.paused = true
	t.pauseAtMs = nowMs
	return true
}

// Resume 恢复计时。缺陷：暂停时长没有累计。
func (t *Timer) Resume(nowMs int64) {
	t.paused = false
}

// Tick 推进一帧。缺陷：不管暂停与否都扣时间，而且只看调用次数。
func (t *Timer) Tick(deltaMs int64) {
	t.ticks++
	t.remainingMs -= deltaMs
}

// Remaining 是剩余时间。缺陷：用的是 Tick 累计值，与 nowMs 无关。
func (t *Timer) Remaining(nowMs int64) int64 { return t.remainingMs }

// Overtime 是否处于加时。缺陷：从来没有加时。
func (t *Timer) Overtime(nowMs int64) bool { return t.overtime }

// Expired 是否已经打完。缺陷：用严格小于判定，剩余正好为 0 不算打完。
func (t *Timer) Expired(nowMs int64) bool { return t.remainingMs < 0 }
