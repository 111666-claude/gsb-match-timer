// Command match-timer 跑对局计时样例。
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
	sample := flags.String("sample", "pause", "pause / expired / idem")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "pause" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		timer.Pause(1000)
		for index := 0; index < 10; index++ {
			timer.Tick(1000)
		}
		fmt.Fprintf(stdout, "remaining=%d\n", timer.Remaining(1000))
		return 0
	}
	if *sample == "expired" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		for index := 0; index < 600; index++ {
			timer.Tick(1000)
		}
		fmt.Fprintf(stdout, "expired=%v\n", timer.Expired(600000))
		return 0
	}
	if *sample == "idem" {
		timer := matchtimer.NewTimer()
		timer.Start(0)
		fmt.Fprintf(stdout, "second=%v\n", timer.Start(100))
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample pause|expired|idem")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
