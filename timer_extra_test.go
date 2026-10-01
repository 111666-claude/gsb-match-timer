package matchtimer

import (
	"math/rand"
	"testing"
)

func TestPauseQuota(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if !timer.Pause(0) {
		t.Fatal("第一次暂停应该被接受")
	}
	timer.Resume(1)
	if !timer.Pause(2) {
		t.Fatal("第二次暂停应该被接受")
	}
	timer.Resume(3)
	if timer.Pause(4) {
		t.Fatal("第三次暂停超出配额，必须返回 false")
	}
	// 被拒绝的暂停不能真正暂停时钟：总暂停时长仍是 2ms。
	if got := timer.Remaining(1000); got != MatchLimitMs-998 {
		t.Fatalf("超额暂停不应改变状态，剩余=%d", got)
	}
}

func TestPauseWhilePausedRejected(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if !timer.Pause(100) {
		t.Fatal("首次暂停应被接受")
	}
	if timer.Pause(200) {
		t.Fatal("已暂停时再请求暂停必须返回 false，且不占用配额")
	}
	timer.Resume(300)
	// 重复暂停没有消耗配额，这仍应是第二次有效暂停。
	if !timer.Pause(400) {
		t.Fatal("重复暂停不应占用配额")
	}
}

func TestRealTimeProgress(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if got := timer.Remaining(1000); got != 599000 {
		t.Fatalf("真实过 1 秒后剩余应为 599000，实际 %d", got)
	}
	timer.Pause(10000)
	timer.Resume(20000) // 暂停 10 秒
	if got := timer.Remaining(30000); got != 580000 {
		t.Fatalf("暂停期间不推进，30 秒时刻剩余应为 580000，实际 %d", got)
	}
}

func TestOvertimeTransitionAndBoundary(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if timer.Overtime(599999) {
		t.Fatal("差 1ms 时不应进入加时")
	}
	if got := timer.Remaining(599999); got != 1 {
		t.Fatalf("边界前剩余应为 1，实际 %d", got)
	}
	if !timer.Overtime(600000) {
		t.Fatal("剩余正好为 0 时应进入加时")
	}
	if got := timer.Remaining(600000); got != OvertimeLimitMs {
		t.Fatalf("刚进加时剩余应为 60000，实际 %d", got)
	}
	if timer.Expired(659999) {
		t.Fatal("加时还差 1ms 时不算打完")
	}
	if !timer.Expired(660000) {
		t.Fatal("加时正好走完（含边界）才算打完")
	}
	if got := timer.Remaining(700000); got != 0 {
		t.Fatalf("打完后剩余下限为 0，实际 %d", got)
	}
}

func TestPauseInvalidInOvertime(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if timer.Pause(600000) {
		t.Fatal("加时期间暂停必须无效（含正好归零的边界）")
	}
	// 被拒绝后时钟仍在走。
	if got := timer.Remaining(610000); got != 50000 {
		t.Fatalf("加时中暂停无效，剩余应为 50000，实际 %d", got)
	}
	if timer.Pause(610000) {
		t.Fatal("加时期间持续无效")
	}
}

func TestClockback(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if got := timer.Remaining(5000); got != 595000 {
		t.Fatalf("过 5 秒剩余应为 595000，实际 %d", got)
	}
	// 时钟回拨到 1 秒，这一秒不能重复扣。
	if got := timer.Remaining(1000); got != 595000 {
		t.Fatalf("回拨后剩余应保持 595000，实际 %d", got)
	}
	if got := timer.Remaining(6000); got != 594000 {
		t.Fatalf("回到正向之后应继续推进，实际 %d", got)
	}
}

func TestClockbackDuringPause(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	timer.Pause(10000)
	// 暂停中曾观测到 20000，随后时钟回拨，恢复时刻不能倒退暂停起点。
	if got := timer.Remaining(20000); got != 590000 {
		t.Fatalf("暂停中不推进，实际 %d", got)
	}
	timer.Resume(15000) // 回拨了 5 秒
	if got := timer.Remaining(25000); got != 585000 {
		t.Fatalf("回拨保护下暂停时长应按高水位算，剩余 585000，实际 %d", got)
	}
}

func TestRepeatedStartDoesNotReset(t *testing.T) {
	timer := NewTimer()
	if !timer.Start(0) {
		t.Fatal("首次开始应返回 true")
	}
	if timer.Start(100000) {
		t.Fatal("重复开始必须返回 false")
	}
	if got := timer.Remaining(100000); got != 500000 {
		t.Fatal("重复开始不得重置开始时刻与进度")
	}
}

func TestReplayDeterministic(t *testing.T) {
	events := []struct {
		kind string
		at   int64
	}{
		{"start", 1000}, {"pause", 5000}, {"resume", 9000},
		{"query", 12000}, {"query", 8000}, {"pause", 20000},
		{"resume", 30000}, {"query", 601000}, {"query", 700000},
	}
	run := func() []int64 {
		timer := NewTimer()
		var out []int64
		for _, event := range events {
			switch event.kind {
			case "start":
				timer.Start(event.at)
			case "pause":
				timer.Pause(event.at)
			case "resume":
				timer.Resume(event.at)
			case "query":
				out = append(out, timer.Remaining(event.at))
			}
		}
		return out
	}
	first := run()
	second := run()
	if len(first) != len(second) {
		t.Fatal("重放结果长度不一致")
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("同一串事件重放结果不一致：%v vs %v", first, second)
		}
	}
}

func TestScale100k(t *testing.T) {
	const n = 100000
	timers := make([]*Timer, n)
	for index := range timers {
		timers[index] = NewTimer()
		timers[index].Start(int64(index))
	}
	// 每秒 10 万次查询量级：每局只存固定字段，单次查询 O(1)。
	for index := range timers {
		now := int64(index) + 1000
		if got := timers[index].Remaining(now); got != MatchLimitMs-1000 {
			t.Fatalf("第 %d 局剩余不正确：%d", index, got)
		}
	}
}

// referenceTimer 是独立写的参考实现：只记录开始时刻和被接受的暂停区间，
// 每次结果都从完整事件历史即时推导。
type referenceTimer struct {
	started    bool
	startMs    int64
	highMs     int64
	pauseCount int
	paused     bool
	pauseStart int64
	intervals  [][2]int64
}

func (r *referenceTimer) played(eff int64) int64 {
	played := eff - r.startMs
	for _, interval := range r.intervals {
		played -= interval[1] - interval[0]
	}
	if r.paused {
		played -= eff - r.pauseStart
	}
	if played < 0 {
		played = 0
	}
	return played
}

func (r *referenceTimer) start(now int64) bool {
	if r.started {
		return false
	}
	r.started = true
	r.startMs = now
	r.highMs = now
	return true
}

func (r *referenceTimer) pause(now int64) bool {
	if !r.started || r.paused || r.pauseCount >= MaxPauses {
		return false
	}
	eff := now
	if eff < r.highMs {
		eff = r.highMs
	}
	if r.played(eff) >= MatchLimitMs {
		return false
	}
	r.highMs = eff
	r.paused = true
	r.pauseStart = eff
	r.pauseCount++
	return true
}

func (r *referenceTimer) resume(now int64) {
	if !r.started || !r.paused {
		return
	}
	eff := now
	if eff < r.highMs {
		eff = r.highMs
	}
	r.intervals = append(r.intervals, [2]int64{r.pauseStart, eff})
	r.paused = false
	r.highMs = eff
}

type refSnapshot struct {
	remaining int64
	overtime  bool
	expired   bool
}

func (r *referenceTimer) query(now int64) refSnapshot {
	if !r.started {
		return refSnapshot{remaining: MatchLimitMs}
	}
	if now > r.highMs {
		r.highMs = now
	}
	played := r.played(r.highMs)
	snap := refSnapshot{overtime: played >= MatchLimitMs, expired: played >= MatchLimitMs+OvertimeLimitMs}
	if played < MatchLimitMs {
		snap.remaining = MatchLimitMs - played
	} else {
		snap.remaining = OvertimeLimitMs - (played - MatchLimitMs)
		if snap.remaining < 0 {
			snap.remaining = 0
		}
	}
	return snap
}

func TestRandomDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20261001))
	for sequence := 0; sequence < 200; sequence++ {
		actual := NewTimer()
		reference := &referenceTimer{}
		now := rng.Int63n(2001) - 1000
		events := 30 + rng.Intn(40)
		for event := 0; event < events; event++ {
			now += rng.Int63n(25001) - 5000 // 允许回拨
			switch rng.Intn(5) {
			case 0:
				got := actual.Start(now)
				want := reference.start(now)
				if got != want {
					t.Fatalf("序列 %d 事件 %d Start(%d)：实际 %v 参考 %v", sequence, event, now, got, want)
				}
			case 1:
				got := actual.Pause(now)
				want := reference.pause(now)
				if got != want {
					t.Fatalf("序列 %d 事件 %d Pause(%d)：实际 %v 参考 %v", sequence, event, now, got, want)
				}
			case 2:
				actual.Resume(now)
				reference.resume(now)
			case 3:
				got := actual.Remaining(now)
				want := reference.query(now)
				if got != want.remaining {
					t.Fatalf("序列 %d 事件 %d Remaining(%d)：实际 %d 参考 %d", sequence, event, now, got, want.remaining)
				}
			default:
				gotO := actual.Overtime(now)
				gotE := actual.Expired(now)
				gotR := actual.Remaining(now)
				want := reference.query(now)
				if gotR != want.remaining || gotO != want.overtime || gotE != want.expired {
					t.Fatalf("序列 %d 事件 %d 状态(%d)：实际 (%d,%v,%v) 参考 (%d,%v,%v)",
						sequence, event, now, gotR, gotO, gotE, want.remaining, want.overtime, want.expired)
				}
			}
		}
	}
}
