// Command line-balancer computes the takt time for a required production rate
// and greedily balances tasks across stations, reporting the bottleneck station
// and overall line efficiency.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"line-balancer/internal/line"
)

func main() {
	in := flag.String("in", "", "tasks CSV file (columns: task_name, seconds) (required)")
	demand := flag.Int("demand", 0, "required units per shift (required, >0)")
	shift := flag.Float64("time", 28800, "available production time in seconds (default 8h = 28800)")
	out := flag.String("out", "", "output file; empty writes to stdout")
	flag.Parse()

	if *in == "" {
		fatal("missing required -in (tasks CSV)")
	}
	if *demand <= 0 {
		fatal("missing or invalid -demand (must be > 0)")
	}
	f, err := os.Open(*in)
	if err != nil {
		fatal("open %q: %v", *in, err)
	}
	defer f.Close()

	tasks, err := line.ParseTasks(f)
	if err != nil {
		fatal("parse %q: %v", *in, err)
	}
	res, err := line.Analyze(tasks, *demand, *shift)
	if err != nil {
		fatal("%v", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Line balance (demand=%d/shift, available=%.0fs)\n", *demand, *shift)
	fmt.Fprintf(&b, "  takt time        : %.2f s/unit\n", res.TaktTime)
	fmt.Fprintf(&b, "  total task time   : %.2f s\n", res.TotalTime)
	fmt.Fprintf(&b, "  stations          : %d\n", res.StationCount)
	fmt.Fprintf(&b, "  bottleneck station: #%d (load %.2f s", res.Bottleneck+1, res.MaxLoad)
	if res.MaxLoad > res.TaktTime {
		fmt.Fprintf(&b, ", EXCEEDS takt!)")
	}
	fmt.Fprintf(&b, ")\n")
	if res.MaxLoad > res.TaktTime {
		fmt.Fprintf(&b, "  line efficiency   : n/a (infeasible — bottleneck exceeds takt)\n")
	} else {
		fmt.Fprintf(&b, "  line efficiency   : %.1f%%\n", res.Efficiency)
	}
	fmt.Fprintf(&b, "\nAssignment:\n")
	for i, s := range res.Stations {
		fmt.Fprintf(&b, "  station #%d (load %.2fs): %s\n", i+1, s.Load, strings.Join(s.Tasks, ", "))
	}

	if *out == "" {
		fmt.Print(b.String())
	} else if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fatal("write %q: %v", *out, err)
	}
}

func fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "line-balancer: "+format+"\n", a...)
	os.Exit(1)
}
