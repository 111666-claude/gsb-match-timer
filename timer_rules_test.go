package matchtimer

import (
	"math/rand"
	"testing"
)

// ---- 暂停配额 ----

func TestPauseQuota(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if !timer.Pause(0) {
		t.Fatal("第一次暂停应该被接受")
	}
	timer.Resume(1000)
	if !timer.Pause(2000) {
		t.Fatal("第二次暂停应该被接受")
	}
	timer.Resume(3000)
	if timer.Pause(4000) {
		t.Fatal("第三次暂停应该被拒绝")
	}
	if timer.Pause(5000) {
		t.Fatal("第四次暂停也应该被拒绝")
	}
	// 被拒绝的暂停不改状态：计时照常推进。
	if got := timer.Remaining(6000); got != MatchLimitMs-4000 {
		t.Fatalf("拒绝暂停后剩余时间错误: %d", got)
	}
}

func TestRejectedPauseKeepsState(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	timer.Pause(0)
	timer.Resume(1000)
	timer.Pause(2000)
	timer.Resume(3000)
	before := timer.Remaining(4000)
	timer.Pause(4000) // 超配额，拒绝
	if got := timer.Remaining(4000); got != before {
		t.Fatalf("拒绝暂停不该改状态: %d -> %d", before, got)
	}
}

func TestPauseFreezesTime(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	timer.Pause(1000)
	if got := timer.Remaining(5000); got != MatchLimitMs-1000 {
		t.Fatalf("暂停期间不该推进: %d", got)
	}
	if got := timer.Remaining(9000); got != MatchLimitMs-1000 {
		t.Fatalf("暂停期间不该推进: %d", got)
	}
	timer.Resume(10000)
	if got := timer.Remaining(11000); got != MatchLimitMs-2000 {
		t.Fatalf("恢复后按真实时间推进: %d", got)
	}
}

func TestPauseWhilePausedRejected(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if !timer.Pause(100) {
		t.Fatal("第一次暂停应该被接受")
	}
	if timer.Pause(200) {
		t.Fatal("暂停中再次暂停应该被拒绝")
	}
}

// ---- 加时赛 ----

func TestOvertimeEntry(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if timer.Overtime(MatchLimitMs - 1) {
		t.Fatal("常规时间未走完不该进加时")
	}
	if !timer.Overtime(MatchLimitMs) {
		t.Fatal("常规时间走完应该进加时")
	}
	if got := timer.Remaining(MatchLimitMs); got != OvertimeLimitMs {
		t.Fatalf("进加时瞬间剩余应为加时全额: %d", got)
	}
}

func TestOvertimeCountsSeparately(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if got := timer.Remaining(MatchLimitMs + 10000); got != OvertimeLimitMs-10000 {
		t.Fatalf("加时单独计时: %d", got)
	}
}

func TestPauseRejectedInOvertime(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if timer.Pause(MatchLimitMs + 1000) {
		t.Fatal("加时期间暂停应该无效")
	}
	if got := timer.Remaining(MatchLimitMs + 2000); got != OvertimeLimitMs-2000 {
		t.Fatalf("加时期间暂停不该影响计时: %d", got)
	}
}

func TestExpiredBoundaries(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if timer.Expired(MatchLimitMs) {
		t.Fatal("刚进加时还不算打完")
	}
	if timer.Expired(MatchLimitMs + OvertimeLimitMs - 1) {
		t.Fatal("加时未走完不算打完")
	}
	if !timer.Expired(MatchLimitMs + OvertimeLimitMs) {
		t.Fatal("加时走完（含边界）应该算打完")
	}
	if !timer.Expired(MatchLimitMs + OvertimeLimitMs + 1) {
		t.Fatal("加时走完后应该算打完")
	}
}

func TestOvertimeWithPauseAccumulated(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	timer.Pause(1000)
	timer.Resume(2000) // 暂停 1 秒
	// 常规时间走完的时刻整体后移 1 秒。
	if timer.Overtime(MatchLimitMs) {
		t.Fatal("暂停时长应顺延加时起点")
	}
	if !timer.Overtime(MatchLimitMs + 1000) {
		t.Fatal("暂停 1 秒后加时应在 601000 开始")
	}
	if got := timer.Remaining(MatchLimitMs + 1000); got != OvertimeLimitMs {
		t.Fatalf("加时起点剩余应为加时全额: %d", got)
	}
}

// ---- 时钟回拨 ----

func TestClockBackDoesNotAdvance(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	if got := timer.Remaining(5000); got != MatchLimitMs-5000 {
		t.Fatalf("推进 5 秒: %d", got)
	}
	if got := timer.Remaining(1000); got != MatchLimitMs-5000 {
		t.Fatalf("回拨不该回退剩余时间: %d", got)
	}
	if got := timer.Remaining(6000); got != MatchLimitMs-6000 {
		t.Fatalf("回拨后继续按真实时间推进: %d", got)
	}
}

func TestClockBackWhilePaused(t *testing.T) {
	timer := NewTimer()
	timer.Start(0)
	timer.Pause(1000)
	if got := timer.Remaining(500); got != MatchLimitMs-1000 {
		t.Fatalf("暂停中回拨不该推进: %d", got)
	}
	timer.Resume(2000)
	if got := timer.Remaining(3000); got != MatchLimitMs-2000 {
		t.Fatalf("恢复后按真实时间推进: %d", got)
	}
}

// ---- 重复开始 ----

func TestRestartRejected(t *testing.T) {
	timer := NewTimer()
	if !timer.Start(100) {
		t.Fatal("第一次开始应该返回 true")
	}
	if timer.Start(200) {
		t.Fatal("重复开始应该返回 false")
	}
	if got := timer.Remaining(1100); got != MatchLimitMs-1000 {
		t.Fatalf("重复开始不该重置进度: %d", got)
	}
}

// ---- 规模：10 万局同时计时 ----

func TestManyTimers(t *testing.T) {
	const count = 100000
	timers := make([]*Timer, count)
	for index := range timers {
		timers[index] = NewTimer()
		timers[index].Start(int64(index % 1000))
	}
	for index, timer := range timers {
		want := MatchLimitMs - (5000 - int64(index%1000))
		if got := timer.Remaining(5000); got != want {
			t.Fatalf("第 %d 局剩余时间错误: got %d want %d", index, got, want)
		}
	}
}

// ---- 对照自证：随机序列与参考实现比对 ----

// oracle 是独立的参考实现：真实时间推进 + 暂停配额 + 加时 + 回拨保护。
type oracle struct {
	startedMs int64
	lastMs    int64
	pausedMs  int64
	pauseAt   int64
	pauses    int
	started   bool
	overtime  bool
}

func newOracle() *oracle { return &oracle{pauseAt: -1} }

func (o *oracle) clamp(now int64) int64 {
	if now < o.lastMs {
		return o.lastMs
	}
	return now
}

func (o *oracle) advance(nowMs int64) {
	if !o.started {
		return
	}
	now := o.clamp(nowMs)
	if o.pauseAt >= 0 {
		if now > o.pauseAt {
			o.pausedMs += now - o.pauseAt
			o.pauseAt = now
		}
	} else if !o.overtime && now-o.startedMs-o.pausedMs >= MatchLimitMs {
		o.overtime = true
	}
	o.lastMs = now
}

func (o *oracle) start(nowMs int64) bool {
	if o.started {
		return false
	}
	o.started = true
	o.startedMs = nowMs
	o.lastMs = nowMs
	return true
}

func (o *oracle) pause(nowMs int64) bool {
	if !o.started || o.overtime || o.pauseAt >= 0 || o.pauses >= MaxPauses {
		return false
	}
	o.advance(nowMs)
	if o.overtime {
		return false
	}
	o.pauses++
	o.pauseAt = o.lastMs
	return true
}

func (o *oracle) resume(nowMs int64) {
	if !o.started || o.pauseAt < 0 {
		return
	}
	o.advance(nowMs)
	o.pauseAt = -1
}

func (o *oracle) remaining(nowMs int64) int64 {
	o.advance(nowMs)
	if !o.started {
		return MatchLimitMs
	}
	elapsed := o.lastMs - o.startedMs - o.pausedMs
	if elapsed < MatchLimitMs {
		return MatchLimitMs - elapsed
	}
	over := OvertimeLimitMs - (elapsed - MatchLimitMs)
	if over < 0 {
		over = 0
	}
	return over
}

func (o *oracle) isOvertime(nowMs int64) bool {
	o.advance(nowMs)
	return o.overtime
}

func (o *oracle) expired(nowMs int64) bool {
	return o.remaining(nowMs) <= 0
}

type event struct {
	kind int // 0=start 1=pause 2=resume 3=remaining 4=overtime 5=expired
	now  int64
}

func TestDifferentialAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(20261002))
	for trial := 0; trial < 200; trial++ {
		timer := NewTimer()
		oracle := newOracle()
		now := rng.Int63n(1000)
		events := make([]event, 0, 300)
		steps := 50 + rng.Intn(250)
		for step := 0; step < steps; step++ {
			// 大部分时间前进，偶尔回拨（含大回拨）。
			if rng.Intn(6) == 0 {
				now -= rng.Int63n(3000)
			} else {
				now += rng.Int63n(20000)
			}
			kind := rng.Intn(6)
			events = append(events, event{kind: kind, now: now})
			switch kind {
			case 0:
				if got, want := timer.Start(now), oracle.start(now); got != want {
					t.Fatalf("trial %d step %d: Start got %v want %v", trial, step, got, want)
				}
			case 1:
				if got, want := timer.Pause(now), oracle.pause(now); got != want {
					t.Fatalf("trial %d step %d: Pause(%d) got %v want %v", trial, step, now, got, want)
				}
			case 2:
				timer.Resume(now)
				oracle.resume(now)
			case 3:
				if got, want := timer.Remaining(now), oracle.remaining(now); got != want {
					t.Fatalf("trial %d step %d: Remaining(%d) got %d want %d", trial, step, now, got, want)
				}
			case 4:
				if got, want := timer.Overtime(now), oracle.isOvertime(now); got != want {
					t.Fatalf("trial %d step %d: Overtime(%d) got %v want %v", trial, step, now, got, want)
				}
			case 5:
				if got, want := timer.Expired(now), oracle.expired(now); got != want {
					t.Fatalf("trial %d step %d: Expired(%d) got %v want %v", trial, step, now, got, want)
				}
			}
		}
		// 不变量：同一串事件重放结果一致。
		replay := NewTimer()
		for _, ev := range events {
			switch ev.kind {
			case 0:
				replay.Start(ev.now)
			case 1:
				replay.Pause(ev.now)
			case 2:
				replay.Resume(ev.now)
			case 3:
				replay.Remaining(ev.now)
			case 4:
				replay.Overtime(ev.now)
			case 5:
				replay.Expired(ev.now)
			}
		}
		final := now + 1000000
		if got, want := replay.Remaining(final), timer.Remaining(final); got != want {
			t.Fatalf("trial %d: 重放剩余时间不一致 got %d want %d", trial, got, want)
		}
		if got, want := replay.Overtime(final), timer.Overtime(final); got != want {
			t.Fatalf("trial %d: 重放加时状态不一致 got %v want %v", trial, got, want)
		}
	}
}

// ---- 每秒 10 万次查询（性能回归护栏） ----

func BenchmarkRemainingQueries(b *testing.B) {
	timer := NewTimer()
	timer.Start(0)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		timer.Remaining(int64(index))
	}
}
