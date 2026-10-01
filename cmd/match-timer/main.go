// Command match-timer 跑对局时钟样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/matchtimer"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("match-timer", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "pause-quota", "pause-quota / overtime / clockback")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "pause-quota" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		timer.Pause(0)
		timer.Resume(1)
		timer.Pause(2)
		timer.Resume(3)
		fmt.Fprintf(stdout, "third=%v\n", timer.Pause(4))
		return 0
	}
	if *sample == "overtime" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		for index := 0; index < 600; index++ {
			timer.Tick(1000)
		}
		fmt.Fprintf(stdout, "overtime=%v remaining=%d\n",
			timer.Overtime(600000), timer.Remaining(600000))
		return 0
	}
	if *sample == "clockback" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		for index := 0; index < 5; index++ {
			timer.Tick(1000)
		}
		// 服务器时钟回拨到 1 秒，按口径这一秒不能重复扣。
		fmt.Fprintf(stdout, "remaining=%d\n", timer.Remaining(1000))
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample pause-quota|overtime|clockback")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
