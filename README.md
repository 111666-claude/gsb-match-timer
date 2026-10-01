# match-timer

对局时钟：按真实时间推进、暂停有配额、剩余归零进入加时、时钟回拨不推进，只用 Go 标准库。

```
go test ./...
go run ./cmd/match-timer --sample pause-quota
go run ./cmd/match-timer --sample overtime
go run ./cmd/match-timer --sample clockback
```

## 口径（README 为准）

- **按真实时间推进**：剩余时间 = 限额 −（`nowMs` − 开始时刻 − 累计暂停时长），暂停期间不推进。
- **暂停配额**：每局最多 `MaxPauses` 次暂停，超出返回 `false` 且不改状态。
- **加时赛**：剩余时间到 0 时进入加时，加时用 `OvertimeLimitMs` 单独计算，
  加时期间暂停无效；加时也走完才算打完（判定含边界）。
- **时钟回拨保护**：`nowMs` 小于上次推进时刻时不推进剩余时间。
- **不变量**：剩余时间只由真实时间、开始时刻、暂停时长决定；同一串事件重放结果一致。
- **规模**：10 万局同时计时、每秒 10 万次查询，单次 O(1)，内存 O(对局数)。

## 输出契约（不改格式）

```
third=false
overtime=true remaining=60000
remaining=599000
```

## 常量

```
MatchLimitMs    = 600000
OvertimeLimitMs = 60000
MaxPauses       = 2
```

## 目录

```
timer.go                对局时钟
cmd/match-timer         命令行入口
timer_test.go           go test 用例
```
