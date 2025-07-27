package types

import (
	"database/sql"
	"encoding/json"
)

type ServerMetric struct {
	Cpu            float32 `json:"cpu"`
	MemoryPercent  float32 `json:"memory_percent"`
	MemoryUsed     uint64  `json:"memory_used"`
	MemoryTotal    uint64  `json:"memory_total"`
	DiskPercent    float32 `json:"disk_percent"`
	DiskUsed       uint64  `json:"disk_used"`
	DiskTotal      uint64  `json:"disk_total"`
	NetworkOutRate uint64  `json:"network_out_rate"`
	NetworkInRate  uint64  `json:"network_in_rate"`
	SwapUsed       uint64  `json:"swap_used"`
	SwapTotal      uint64  `json:"swap_total"`
	SwapPercent    float32 `json:"swap_percent"`
	Uptime         uint64  `json:"uptime"`
	Load1          float32 `json:"load_1"`
	Load5          float32 `json:"load_5"`
	Load15         float32 `json:"load_15"`
}

type ServerInfo struct {
	Cpu     string `json:"cpu"`
	Arch    string `json:"arch"`
	OS      string `json:"operating_system"`
	Version string `json:"version"`
}

type ServiceMetric struct {
	Timestamp uint64    `json:"timestamp"`
	To        int32     `json:"to"`
	Min       NullInt64 `json:"min"`
	Max       NullInt64 `json:"max"`
	Avg       NullInt64 `json:"avg"`
	Median    NullInt64 `json:"median"`
	Loss      float32   `json:"loss"`
}

type Service struct {
	Id    int32  `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Host  string `json:"host"`
}

// Nullable Int64 that overrides sql.NullInt64
type NullInt64 struct {
	sql.NullInt64
}

func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

func (ni *NullInt64) UnmarshalJSON(data []byte) error {
	var i *int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Valid = true
		ni.Int64 = *i
	} else {
		ni.Valid = false
	}
	return nil
}
