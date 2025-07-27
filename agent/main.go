package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/sduoduo233/pbb/controllers/types"
	"github.com/sduoduo233/pbb/update"
)

var URL string
var SECRET string

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("could not load .env file", "err", err)
	}

	URL = os.Getenv("AGENT_URL")
	SECRET = os.Getenv("AGENT_SECRET")

	if os.Getenv("DEBUG") == "1" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Warn("agent")

	slog.Warn("reporting to", "url", URL)

	// auto update
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	s.NewJob(gocron.CronJob("0 3 * * *", false), gocron.NewTask(func() {
		err := update.AutoUpdate("agent")
		if err != nil {
			slog.Error("auto update error", "error", err)
		}
	}))
	s.Start()

	go func() {
		// report system info every 10 minutes
		for {
			err := reportSystemInfo()
			if err != nil {
				slog.Error("could not report system info", "err", err)
				time.Sleep(time.Second * 30)
				continue
			}

			time.Sleep(time.Minute * 10)
		}
	}()

	go func() {
		// report ping every 5 minutes
		ticker := time.NewTicker(time.Minute * 5)
		for {
			err := pingServices()
			if err != nil {
				slog.Error("could not report ping", "err", err)
			}
			<-ticker.C
		}
	}()

	for {

		m, err := getMetirc()
		if err != nil {
			slog.Error("could not get metrics", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		_, err = sendReport(URL+"/metric", m)
		if err != nil {
			slog.Error("could not send request", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		time.Sleep(time.Second * 5)

	}
}

func sendReport(url string, m any) ([]byte, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("x-secret", SECRET)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response: status=%s, body=%s", resp.Status, string(body))
	}

	return body, nil
}

var (
	lastTime      time.Time = time.Now()
	lastIoCounter *net.IOCountersStat
)

func getMetirc() (*types.ServerMetric, error) {
	m := types.ServerMetric{}

	// cpu

	cpuPercent, err := cpu.Percent(time.Millisecond*100, false)
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	m.Cpu = float32(cpuPercent[0])

	// disk

	diskUsage, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	m.DiskPercent = float32(diskUsage.UsedPercent)
	m.DiskTotal = uint64(diskUsage.Total)
	m.DiskUsed = uint64(diskUsage.Used)

	// network

	ioCounters, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("net: %w", err)
	}
	ioCounter := ioCounters[0]

	if lastIoCounter != nil {
		m.NetworkInRate = (uint64(ioCounter.BytesRecv) - uint64(lastIoCounter.BytesRecv)) / uint64(time.Since(lastTime).Seconds())
		m.NetworkOutRate = (uint64(ioCounter.BytesSent) - uint64(lastIoCounter.BytesSent)) / uint64(time.Since(lastTime).Seconds())
	}
	lastIoCounter = &ioCounter
	lastTime = time.Now()

	// memory

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("mem: %w", err)
	}

	m.MemoryPercent = float32(memoryStat.UsedPercent)
	m.MemoryTotal = uint64(memoryStat.Total)
	m.MemoryUsed = uint64(memoryStat.Used)

	// swap

	swapStat, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("swap: %w", err)
	}

	m.SwapUsed = uint64(swapStat.Used)
	m.SwapTotal = uint64(swapStat.Total)
	m.SwapPercent = float32(swapStat.UsedPercent)

	// uptime

	uptime, err := host.Uptime()
	if err != nil {
		return nil, fmt.Errorf("uptime: %w", err)

	}
	m.Uptime = uptime

	// load

	avgStat, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("load avg: %w", err)
	}
	m.Load1 = float32(avgStat.Load1)
	m.Load5 = float32(avgStat.Load5)
	m.Load15 = float32(avgStat.Load15)

	return &m, nil
}

func reportSystemInfo() error {
	var s types.ServerInfo

	s.Version = update.CURRENT_VERSION

	// cpu

	cpuStat, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("cpu: %w", err)
	}

	cpuCores := make(map[string]int)
	for _, c := range cpuStat {
		_, ok := cpuCores[c.ModelName]
		if ok {
			cpuCores[c.ModelName] += int(c.Cores)
		} else {
			cpuCores[c.ModelName] = int(c.Cores)
		}
	}

	for k, v := range cpuCores {
		s.Cpu += k
		s.Cpu += " (" + strconv.Itoa(v) + " Cores)"
		s.Cpu += ","
	}
	s.Cpu = s.Cpu[:len(s.Cpu)-1]

	// arch

	arch, err := host.KernelArch()
	if err != nil {
		return fmt.Errorf("arch: %w", err)
	}
	s.Arch = arch

	// os

	platform, _, version, err := host.PlatformInformation()
	if err != nil {
		return fmt.Errorf("os: %w", err)
	}

	s.OS = fmt.Sprintf("%s %s", platform, version)

	// send report

	_, err = sendReport(URL+"/info", &s)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	return nil
}

var relatedServices = make([]types.Service, 0)
var serviceMetrics = make([]types.ServiceMetric, 0)

func pingServices() error {
	serviceMetrics = make([]types.ServiceMetric, 0)

	var resultsCh = make(chan types.ServiceMetric)

	for _, s := range relatedServices {
		timestamp := time.Now().Unix()
		timestamp = timestamp / 300 * 300

		switch s.Type {
		case "ping":
			go icmpPing(resultsCh, s.Host, uint64(timestamp), s.Id)
		case "tcp":
			go tcpPing(resultsCh, s.Host, uint64(timestamp), s.Id)
		}
	}

	for range len(relatedServices) {
		serviceMetrics = append(serviceMetrics, <-resultsCh)
	}

	body, err := sendReport(URL+"/service", serviceMetrics)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	err = json.Unmarshal(body, &relatedServices)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	return nil
}
