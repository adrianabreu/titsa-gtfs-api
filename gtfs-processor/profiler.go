package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/pterm/pterm"
)

type ProfileStats struct {
	startTime    time.Time
	endTime      time.Time
	minMemory    uint64
	maxMemory    uint64
	totalMemory  uint64
	samples      int
	startCPUTime time.Duration
	enabled      bool
	stopChan     chan bool
}

func NewProfiler() *ProfileStats {
	return &ProfileStats{
		minMemory: ^uint64(0), // Max uint64
		enabled:   false,
		stopChan:  make(chan bool),
	}
}

func (p *ProfileStats) Start() {
	if !profileEnabled {
		return
	}

	p.enabled = true
	p.startTime = time.Now()

	// Force GC to get accurate starting stats
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	p.minMemory = m.Alloc
	p.maxMemory = m.Alloc

	// Start monitoring goroutine
	go p.monitor()
}

func (p *ProfileStats) monitor() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			if m.Alloc < p.minMemory {
				p.minMemory = m.Alloc
			}
			if m.Alloc > p.maxMemory {
				p.maxMemory = m.Alloc
			}

			p.totalMemory += m.Alloc
			p.samples++

		case <-p.stopChan:
			return
		}
	}
}

func (p *ProfileStats) Stop() {
	if !p.enabled {
		return
	}

	p.endTime = time.Now()
	p.stopChan <- true

	// Final GC and stats
	runtime.GC()

	p.printStats()
}

func (p *ProfileStats) printStats() {
	duration := p.endTime.Sub(p.startTime)

	avgMemory := uint64(0)
	if p.samples > 0 {
		avgMemory = p.totalMemory / uint64(p.samples)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pterm.Println()
	pterm.DefaultSection.Println("Profile Statistics")

	// Performance metrics table
	tableData := pterm.TableData{
		{"Metric", "Value"},
		{"Duration", formatDuration(duration)},
		{"Min Memory", formatBytes(p.minMemory)},
		{"Max Memory", formatBytes(p.maxMemory)},
		{"Avg Memory", formatBytes(avgMemory)},
		{"Final Memory", formatBytes(m.Alloc)},
		{"Total Allocated", formatBytes(m.TotalAlloc)},
		{"Sys Memory", formatBytes(m.Sys)},
		{"GC Runs", fmt.Sprintf("%d", m.NumGC)},
		{"Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())},
		{"CPU Cores", fmt.Sprintf("%d", runtime.NumCPU())},
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	// Memory breakdown
	pterm.Println()
	pterm.Info.Println("Memory Breakdown:")
	pterm.Printf("  Heap Alloc:    %s\n", formatBytes(m.HeapAlloc))
	pterm.Printf("  Heap Sys:      %s\n", formatBytes(m.HeapSys))
	pterm.Printf("  Heap In Use:   %s\n", formatBytes(m.HeapInuse))
	pterm.Printf("  Stack In Use:  %s\n", formatBytes(m.StackInuse))
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
	return fmt.Sprintf("%.2f min", d.Minutes())
}
