# match-timer

对局计时：按真实时间推进、暂停不计时、超时含边界、开始幂等，只用 Go 标准库。

```
go test ./...
go run ./cmd/match-timer --sample pause
go run ./cmd/match-timer --sample expired
go run ./cmd/match-timer --sample idem
```

## 口径（README 为准）

- **暂停不计时**：暂停期间不论推进多少帧，`Remaining` 都不减少；恢复后把暂停时长从已用时间里扣掉。
- **超时含边界**：剩余时间小于等于 0 就算超时。
- **开始幂等**：已经开始的计时器再次 `Start` 返回 `false`，并且不重置进度。
- **不变量**：`Remaining` 永远不因为暂停或重复推进而掉到 0 以下再回弹；同一串操作重复执行结果相同。
- **规模**：单服 10 万局同时计时、每秒 10 万次查询，单次 O(1)，内存 O(对局数)。

## 输出契约（不改格式）

```
remaining=599000
expired=true
second=false
```

## 常量

```
MatchLimitMs = 600000
```

## 目录

```
timer.go                对局计时
cmd/match-timer         命令行入口
timer_test.go           go test 用例
```
