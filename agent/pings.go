package main

import (
	"database/sql"
	"log/slog"
	"net"
	"slices"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/sduoduo233/pbb/controllers/types"
	"github.com/sduoduo233/pbb/db"
)

func icmpPing(ch chan types.ServiceMetric, host string, timestamp uint64, to int32) {
	r := types.ServiceMetric{
		Min:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Max:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Avg:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Median:    db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Loss:      1,
		Timestamp: timestamp,
		To:        to,
	}
	defer func() {
		ch <- r
	}()

	pinger := probing.New(host)
	pinger.Count = 20
	pinger.Interval = time.Second
	pinger.Size = 56
	pinger.ResolveTimeout = time.Second * 5
	pinger.Timeout = time.Minute
	pinger.RecordRtts = true

	err := pinger.Resolve()
	if err != nil {
		slog.Debug("ping resolve error", "err", err, "host", host)
		return
	}

	err = pinger.Run()
	if err != nil {
		slog.Debug("ping error", "err", err, "host", host)
		return
	}

	stats := pinger.Statistics()
	if len(stats.Rtts) == 0 {
		return
	}

	r.Loss = float32(1) - float32(stats.PacketsRecv)/float32(stats.PacketsSent)

	rtts := make([]int64, 0, len(stats.Rtts))
	for _, rtt := range stats.Rtts {
		rtts = append(rtts, rtt.Microseconds())
	}
	slices.Sort(rtts)

	r.Min.Valid = true
	r.Min.Int64 = rtts[0]
	r.Max.Valid = true
	r.Max.Int64 = rtts[len(rtts)-1]
	r.Avg.Valid = true
	r.Avg.Int64 = stats.AvgRtt.Microseconds()
	r.Median.Valid = true
	r.Median.Int64 = rtts[len(rtts)/2]
}

func tcpPing(ch chan types.ServiceMetric, host string, timestamp uint64, to int32) {
	r := types.ServiceMetric{
		Min:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Max:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Avg:       db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Median:    db.NullInt64{NullInt64: sql.NullInt64{Valid: false}},
		Loss:      1,
		Timestamp: timestamp,
		To:        to,
	}
	defer func() {
		ch <- r
	}()

	rtts := make([]int64, 0, 5)

	for range 5 {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", host, time.Second*10)
		if err != nil {
			slog.Debug("tcp ping dial error", "err", err, "host", host)
			continue
		}
		err = conn.Close()
		if err != nil {
			slog.Debug("tcp ping close error", "err", err, "host", host)
		}
		rtts = append(rtts, time.Since(start).Microseconds())
	}

	if len(rtts) == 0 {
		return
	}

	slices.Sort(rtts)

	var sum int64
	for _, r := range rtts {
		sum += r
	}

	r.Min.Valid = true
	r.Min.Int64 = rtts[0]
	r.Max.Valid = true
	r.Max.Int64 = rtts[len(rtts)-1]
	r.Median.Valid = true
	r.Median.Int64 = rtts[len(rtts)/2]
	r.Avg.Valid = true
	r.Avg.Int64 = sum / int64(len(rtts))

	r.Loss = float32(1) - float32(len(rtts))/float32(5)
}
