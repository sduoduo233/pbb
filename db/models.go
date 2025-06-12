package db

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
	Id         int32  `db:"id"`
	Label      string `db:"label"`
	Hidden     bool   `db:"hidden"`
	GroupId    int32  `db:"group_id"`
	LastReport int64  `db:"last_report"`
}
