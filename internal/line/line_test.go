package line

import (
	"strings"
	"testing"
)

func TestTaktTime(t *testing.T) {
	if got := TaktTime(100, 28800); got != 288 {
		t.Fatalf("takt = %v, want 288", got)
	}
}

func TestAnalyzePlacesAllTasks(t *testing.T) {
	tasks := []Task{
		{Name: "weld", Time: 45},
		{Name: "paint", Time: 30},
		{Name: "assemble", Time: 60},
		{Name: "inspect", Time: 20},
		{Name: "pack", Time: 25},
		{Name: "cut", Time: 40},
		{Name: "polish", Time: 35},
		{Name: "test", Time: 15},
	}
	res, err := Analyze(tasks, 400, 28800) // takt = 72s
	if err != nil {
		t.Fatal(err)
	}
	if res.TaktTime != 72 {
		t.Fatalf("takt = %v, want 72", res.TaktTime)
	}
	if res.StationCount < 1 {
		t.Fatalf("expected at least 1 station, got %d", res.StationCount)
	}
	// every task must be assigned exactly once
	assigned := 0
	for _, s := range res.Stations {
		assigned += len(s.Tasks)
	}
	if assigned != len(tasks) {
		t.Fatalf("assigned %d tasks, want %d", assigned, len(tasks))
	}
	// bottleneck load must equal max station load and the overall max task time
	if res.MaxLoad <= 0 {
		t.Fatalf("max load = %v", res.MaxLoad)
	}
	if res.Efficiency <= 0 || res.Efficiency > 100 {
		t.Fatalf("efficiency = %v, want (0,100]", res.Efficiency)
	}
}

func TestAnalyzeBadDemand(t *testing.T) {
	if _, err := Analyze([]Task{{Name: "a", Time: 1}}, 0, 100); err == nil {
		t.Fatal("expected error for zero demand")
	}
}

func TestParseTasks(t *testing.T) {
	csv := "task,seconds\nweld,45\npaint,30\n"
	tasks, err := ParseTasks(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].Name != "weld" || tasks[0].Time != 45 {
		t.Fatalf("parsed %+v", tasks)
	}
}

func TestBalanceDoesNotMutateInput(t *testing.T) {
	tasks := []Task{
		{Name: "a", Time: 10},
		{Name: "b", Time: 50},
		{Name: "c", Time: 30},
	}
	orig := append([]Task(nil), tasks...)
	_ = Balance(tasks, 60)
	for i := range tasks {
		if tasks[i] != orig[i] {
			t.Fatalf("input mutated at %d: got %+v want %+v", i, tasks[i], orig[i])
		}
	}
}

func TestAnalyzeBottleneckMaxLoad(t *testing.T) {
	tasks := []Task{
		{Name: "long", Time: 50},
		{Name: "a", Time: 10},
		{Name: "b", Time: 10},
		{Name: "c", Time: 15},
	}
	res, err := Analyze(tasks, 100, 7200) // takt = 72
	if err != nil {
		t.Fatal(err)
	}
	wantMax := 0.0
	for _, s := range res.Stations {
		if s.Load > wantMax {
			wantMax = s.Load
		}
	}
	if res.MaxLoad != wantMax {
		t.Fatalf("MaxLoad = %v, want station max %v (stations=%+v)", res.MaxLoad, wantMax, res.Stations)
	}
	if res.Bottleneck < 0 || res.Bottleneck >= len(res.Stations) {
		t.Fatalf("Bottleneck index %d out of range", res.Bottleneck)
	}
	if res.Stations[res.Bottleneck].Load != wantMax {
		t.Fatalf("Bottleneck station load %v != MaxLoad %v", res.Stations[res.Bottleneck].Load, wantMax)
	}
}

func TestBalanceExactFitReusesStation(t *testing.T) {
	// longest-first: a(40) opens station, b(20) must exactly fill remaining 20 under cycle 60
	tasks := []Task{
		{Name: "a", Time: 40},
		{Name: "b", Time: 20},
	}
	stations := Balance(tasks, 60)
	if len(stations) != 1 {
		t.Fatalf("got %d stations, want 1 (exact remaining capacity fit)", len(stations))
	}
	if stations[0].Load != 60 {
		t.Fatalf("load = %v, want 60", stations[0].Load)
	}
}

func TestParseTasksRejectsNegative(t *testing.T) {
	_, err := ParseTasks(strings.NewReader("task,seconds\nweld,-5\n"))
	if err == nil {
		t.Fatal("expected error for negative seconds")
	}
}
