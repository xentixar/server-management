// Package metrics samples host CPU, memory, disk, and load for heartbeats.
package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"agent/internal/protocol"
)

// Collector keeps prior /proc/stat counters so CPU % can be derived between samples.
type Collector struct {
	mu   sync.Mutex
	prev *cpuSample
}

type cpuSample struct {
	idle  uint64
	total uint64
}

// Sample reads host metrics. Missing sources are omitted (fail soft).
func (c *Collector) Sample() *protocol.HostMetrics {
	out := &protocol.HostMetrics{}
	c.sampleCPU(out)
	sampleMem(out)
	sampleLoad(out)
	sampleDisk(out)
	if out.CPUPercent == nil && out.MemTotalBytes == nil && out.DiskTotalBytes == nil &&
		out.Load1 == nil && out.Load5 == nil && out.Load15 == nil {
		return nil
	}
	return out
}

func (c *Collector) sampleCPU(out *protocol.HostMetrics) {
	cur, err := readCPUSample()
	if err != nil {
		return
	}
	c.mu.Lock()
	prev := c.prev
	c.prev = &cur
	c.mu.Unlock()
	if prev == nil {
		return
	}
	dtotal := cur.total - prev.total
	didle := cur.idle - prev.idle
	if dtotal == 0 {
		return
	}
	pct := 100 * (1 - float64(didle)/float64(dtotal))
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	out.CPUPercent = &pct
}

func readCPUSample() (cpuSample, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuSample{}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return cpuSample{}, fmt.Errorf("empty /proc/stat")
	}
	return parseCPUStatLine(sc.Text())
}

func parseCPUStatLine(line string) (cpuSample, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, fmt.Errorf("bad cpu line")
	}
	var vals []uint64
	for _, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return cpuSample{}, err
		}
		vals = append(vals, v)
	}
	var total uint64
	for _, v := range vals {
		total += v
	}
	// idle = idle + iowait when present
	idle := vals[3]
	if len(vals) > 4 {
		idle += vals[4]
	}
	return cpuSample{idle: idle, total: total}, nil
}

func sampleMem(out *protocol.HostMetrics) {
	total, avail, err := readMeminfo()
	if err != nil || total == 0 {
		return
	}
	used := total - avail
	if avail > total {
		used = 0
	}
	out.MemTotalBytes = &total
	out.MemUsedBytes = &used
}

func readMeminfo() (total, available uint64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	var memTotal, memAvailable, memFree, buffers, cached uint64
	var sawTotal, sawAvail bool
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		key, kib, ok := parseMeminfoLine(line)
		if !ok {
			continue
		}
		switch key {
		case "MemTotal":
			memTotal = kib * 1024
			sawTotal = true
		case "MemAvailable":
			memAvailable = kib * 1024
			sawAvail = true
		case "MemFree":
			memFree = kib * 1024
		case "Buffers":
			buffers = kib * 1024
		case "Cached":
			cached = kib * 1024
		}
	}
	if err := sc.Err(); err != nil {
		return 0, 0, err
	}
	if !sawTotal {
		return 0, 0, fmt.Errorf("MemTotal missing")
	}
	if !sawAvail {
		memAvailable = memFree + buffers + cached
	}
	return memTotal, memAvailable, nil
}

func parseMeminfoLine(line string) (key string, kib uint64, ok bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	key = strings.TrimSpace(parts[0])
	fields := strings.Fields(parts[1])
	if len(fields) < 1 {
		return "", 0, false
	}
	v, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return "", 0, false
	}
	return key, v, true
}

func sampleLoad(out *protocol.HostMetrics) {
	l1, l5, l15, err := readLoadavg()
	if err != nil {
		return
	}
	out.Load1 = &l1
	out.Load5 = &l5
	out.Load15 = &l15
}

func readLoadavg() (l1, l5, l15 float64, err error) {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	return parseLoadavg(string(raw))
}

func parseLoadavg(s string) (l1, l5, l15 float64, err error) {
	fields := strings.Fields(s)
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("bad loadavg")
	}
	l1, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	l5, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	l15, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return l1, l5, l15, nil
}

func sampleDisk(out *protocol.HostMetrics) {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return
	}
	// Blocks are filesystem block size units; Bavail is free for non-root.
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	if free > total {
		free = total
	}
	used := total - free
	out.DiskTotalBytes = &total
	out.DiskUsedBytes = &used
}
