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

	Arch sql.NullString `db:"arch"`
	OS   sql.NullString `db:"operating_system"`
	Cpu  sql.NullString `db:"cpu"`
}

type ServerWithGroupLabel struct {
	Server
	GroupLabel sql.NullString `db:"group_label"`
}

type ServerMetrics struct {
	Id        int32 `db:"id"`
	CreatedAt int32 `db:"created_at"`
	ServerId  int32 `db:"server_id"`

	Cpu            float32 `db:"cpu"`
	MemoryPercent  float32 `db:"memory_percent"`
	MemoryUsed     int32   `db:"memory_used"`
	MemoryTotal    int32   `db:"memory_total"`
	DiskPercent    float32 `db:"disk_percent"`
	DiskUsed       int32   `db:"disk_used"`
	DiskTotal      int32   `db:"disk_total"`
	NetworkOutRate int32   `db:"network_out_rate"`
	NetworkInRate  int32   `db:"network_in_rate"`
	SwapUsed       int32   `db:"swap_used"`
	SwapTotal      int32   `db:"spaw_total"`
}
