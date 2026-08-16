// Package line models a simple production-line balancing problem: given a set
// of tasks (each with a processing time) and a required production rate, it
// computes the takt time and greedily assigns tasks to stations, then reports
// the bottleneck station and line efficiency.
package line

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Task is one workstation operation with its processing time in seconds.
type Task struct {
	Name string
	Time float64
}

// Station is a work station holding one or more tasks.
type Station struct {
	Tasks []string
	Load  float64
}

// Result summarizes a line-balance analysis.
type Result struct {
	TaktTime     float64
	CycleTime    float64
	Stations     []Station
	StationCount int
	Bottleneck   int // index of the station with the highest load
	MaxLoad      float64
	Efficiency   float64 // percent of available cycle time actually used
	TotalTime    float64
}

// TaktTime is the maximum time per unit allowed to meet demand.
func TaktTime(demand int, availableSec float64) float64 {
	return availableSec / float64(demand)
}

// Balance assigns tasks to stations using a longest-task-first, best-fit
// heuristic: each task goes into the existing station with the smallest
// remaining capacity that still fits; otherwise a new station is opened.
// Tasks whose time exceeds cycleTime get their own (overloaded) station.
func Balance(tasks []Task, cycleTime float64) []Station {
	sorted := append([]Task(nil), tasks...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Time != sorted[j].Time {
			return sorted[i].Time > sorted[j].Time
		}
		return sorted[i].Name < sorted[j].Name
	})

	var stations []Station
	for _, t := range sorted {
		best, bestRem := -1, cycleTime+1
		for i := range stations {
			rem := cycleTime - stations[i].Load
			if rem >= t.Time && rem < bestRem {
				best, bestRem = i, rem
			}
		}
		if best >= 0 {
			stations[best].Tasks = append(stations[best].Tasks, t.Name)
			stations[best].Load += t.Time
		} else {
			stations = append(stations, Station{Tasks: []string{t.Name}, Load: t.Time})
		}
	}
	return stations
}

// Analyze computes takt time, balances the line, and derives the bottleneck and
// efficiency. availableSec is the total available production time (e.g. 28800
// for an 8-hour shift in seconds).
func Analyze(tasks []Task, demand int, availableSec float64) (Result, error) {
	if demand <= 0 {
		return Result{}, fmt.Errorf("demand must be > 0, got %d", demand)
	}
	if availableSec <= 0 {
		return Result{}, fmt.Errorf("available time must be > 0, got %v", availableSec)
	}
	takt := TaktTime(demand, availableSec)
	stations := Balance(tasks, takt)

	total := 0.0
	for _, t := range tasks {
		total += t.Time
	}
	maxLoad, bn := 0.0, 0
	for i, s := range stations {
		if s.Load > maxLoad {
			maxLoad, bn = s.Load, i
		}
	}
	eff := 0.0
	if len(stations) > 0 && takt > 0 {
		eff = total / (float64(len(stations)) * takt) * 100
	}
	return Result{
		TaktTime:     takt,
		CycleTime:    takt,
		Stations:     stations,
		StationCount: len(stations),
		Bottleneck:   bn,
		MaxLoad:      maxLoad,
		Efficiency:   eff,
		TotalTime:    total,
	}, nil
}

// ParseTasks reads a CSV with columns: task_name, seconds (optional header).
func ParseTasks(r io.Reader) ([]Task, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(recs))
	for ri, rec := range recs {
		if ri == 0 && len(rec) > 0 {
			rec[0] = strings.TrimPrefix(rec[0], "\ufeff")
		}
		if len(rec) > 0 {
			h := strings.TrimSpace(strings.ToLower(rec[0]))
			if h == "task" || h == "name" || h == "工序" {
				continue // skip CSV header
			}
		}
		if len(rec) < 2 {
			if len(rec) == 0 || strings.TrimSpace(rec[0]) == "" {
				continue
			}
			return nil, fmt.Errorf("line %d: expected 'name,seconds', got %q", ri+1, strings.Join(rec, ","))
		}
		name := strings.TrimSpace(rec[0])
		t, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid seconds %q", ri+1, rec[1])
		}
		if t < 0 {
			return nil, fmt.Errorf("line %d: negative time %v", ri+1, t)
		}
		tasks = append(tasks, Task{Name: name, Time: t})
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}
	return tasks, nil
}
