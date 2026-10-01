package matchtimer

import "testing"

func TestNewTimerIsFull(t *testing.T) {
	if NewTimer().Remaining(0) != MatchLimitMs {
		t.Fatal("新建时钟应该是满的")
	}
}

func TestNotExpiredInitially(t *testing.T) {
	if NewTimer().Expired(0) {
		t.Fatal("刚开始不该算打完")
	}
}

func TestStartReturnsTrue(t *testing.T) {
	if !NewTimer().Start(0) {
		t.Fatal("第一次开始应该返回 true")
	}
}

func TestConstants(t *testing.T) {
	if MatchLimitMs != 600000 || OvertimeLimitMs != 60000 || MaxPauses != 2 {
		t.Fatal("对局时钟常量被改了")
	}
}
