package monitor

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/gopsutil/cpu"
	"github.com/ikidev/lightning/internal/gopsutil/load"
	"github.com/ikidev/lightning/internal/gopsutil/mem"
	"github.com/ikidev/lightning/internal/gopsutil/net"
	"github.com/ikidev/lightning/internal/gopsutil/process"
)

type stats struct {
	PID statsPID `json:"pid"`
	OS  statsOS  `json:"os"`
}

type statsPID struct {
	CPU   float64 `json:"cpu"`
	RAM   uint64  `json:"ram"`
	Conns int     `json:"conns"`
}

type statsOS struct {
	CPU      float64 `json:"cpu"`
	RAM      uint64  `json:"ram"`
	TotalRAM uint64  `json:"total_ram"`
	LoadAvg  float64 `json:"load_avg"`
	Conns    int     `json:"conns"`
}

var (
	monitPidCpu   atomic.Value
	monitPidRam   atomic.Value
	monitPidConns atomic.Value

	monitOsCpu      atomic.Value
	monitOsRam      atomic.Value
	monitOsTotalRam atomic.Value
	monitOsLoadAvg  atomic.Value
	monitOsConns    atomic.Value
)

var (
	mutex sync.RWMutex
	once  sync.Once
	data  = &stats{}
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Start routine to update statistics
	once.Do(func() {
		p, _ := process.NewProcess(int32(os.Getpid()))

		updateStatistics(p)

		go func() {
			for {
				updateStatistics(p)

				time.Sleep(1 * time.Second)
			}
		}()
	})

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		if req.Method() != lightning.MethodGet {
			return lightning.ErrMethodNotAllowed
		}
		if req.Header.Get(lightning.HeaderAccept) == lightning.MIMEApplicationJSON || cfg.APIOnly {
			mutex.Lock()
			data.PID.CPU = monitPidCpu.Load().(float64)
			data.PID.RAM = monitPidRam.Load().(uint64)
			data.PID.Conns = monitPidConns.Load().(int)

			data.OS.CPU = monitOsCpu.Load().(float64)
			data.OS.RAM = monitOsRam.Load().(uint64)
			data.OS.TotalRAM = monitOsTotalRam.Load().(uint64)
			data.OS.LoadAvg = monitOsLoadAvg.Load().(float64)
			data.OS.Conns = monitOsConns.Load().(int)
			mutex.Unlock()
			return res.Status(lightning.StatusOK).JSON(data)
		}
		res.Header.SetContentType(lightning.MIMETextHTMLCharsetUTF8)
		return res.Status(lightning.StatusOK).Bytes(index)
	}
}

func updateStatistics(p *process.Process) {
	pidCpu, _ := p.CPUPercent()
	monitPidCpu.Store(pidCpu / 10)

	if osCpu, _ := cpu.Percent(0, false); len(osCpu) > 0 {
		monitOsCpu.Store(osCpu[0])
	}

	if pidMem, _ := p.MemoryInfo(); pidMem != nil {
		monitPidRam.Store(pidMem.RSS)
	}

	if osMem, _ := mem.VirtualMemory(); osMem != nil {
		monitOsRam.Store(osMem.Used)
		monitOsTotalRam.Store(osMem.Total)
	}

	if loadAvg, _ := load.Avg(); loadAvg != nil {
		monitOsLoadAvg.Store(loadAvg.Load1)
	}

	pidConns, _ := net.ConnectionsPid("tcp", p.Pid)
	monitPidConns.Store(len(pidConns))

	osConns, _ := net.Connections("tcp")
	monitOsConns.Store(len(osConns))
}
