package db

import "database/sql"

type User struct {
	Id       int32  `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type Token struct {
	Id        int32  `db:"id"`
	Token     string `db:"token"`
	UserId    int32  `db:"user_id"`
	CreatedAt int64  `db:"created_at"`
}

type Group struct {
	Id     int32  `db:"id"`
	Label  string `db:"label"`
	Hidden bool   `db:"hidden"`
}

type Server struct {
	Id         int32         `db:"id"`
	Label      string        `db:"label"`
	Hidden     bool          `db:"hidden"`
	GroupId    sql.NullInt32 `db:"group_id"`
	LastReport sql.NullInt64 `db:"last_report"`
	Secret     string        `db:"secret"`

	Arch    sql.NullString `db:"arch"`
	OS      sql.NullString `db:"operating_system"`
	Cpu     sql.NullString `db:"cpu"`
	Version sql.NullString `db:"version"`
}

type ServerWithGroupLabel struct {
	Server
	GroupLabel sql.NullString `db:"group_label"`
}

type ServerMetrics struct {
	Id        int32  `db:"id" json:"id"`
	CreatedAt uint64 `db:"created_at" json:"created_at"`
	ServerId  int32  `db:"server_id" json:"server_id"`

	Cpu            float32 `db:"cpu" json:"cpu"`
	MemoryPercent  float32 `db:"memory_percent" json:"memory_percent"`
	MemoryUsed     uint64  `db:"memory_used" json:"memory_used"`
	MemoryTotal    uint64  `db:"memory_total" json:"memory_total"`
	DiskPercent    float32 `db:"disk_percent" json:"disk_percent"`
	DiskUsed       uint64  `db:"disk_used" json:"disk_used"`
	DiskTotal      uint64  `db:"disk_total" json:"disk_total"`
	NetworkOutRate uint64  `db:"network_out_rate" json:"network_out_rate"`
	NetworkInRate  uint64  `db:"network_in_rate" json:"network_in_rate"`
	SwapUsed       uint64  `db:"swap_used" json:"swap_used"`
	SwapTotal      uint64  `db:"swap_total" json:"swap_total"`
	SwapPercent    float32 `db:"swap_percent" json:"swap_percent"`
	Uptime         uint64  `db:"uptime" json:"uptime"`
	Load1          float32 `db:"load_1" json:"load_1"`
	Load5          float32 `db:"load_5" json:"load_5"`
	Load15         float32 `db:"load_15" json:"load_15"`
}

type Setting struct {
	Id    int32  `db:"id"`
	Key   string `db:"key"`
	Value string `db:"value"`
}

const (
	IncidentStateOngoing  = "ongoing"
	IncidentStateResolved = "resolved"
)

type Incident struct {
	Id        int32         `db:"id"`
	ServerId  int32         `db:"server_id"`
	StartedAt int64         `db:"started_at"`
	EndedAt   sql.NullInt64 `db:"ended_at"`
	State     string        `db:"state"`
}
