package models

type Session struct {
	Sid        string
	Token      string
	CreatedAt  int64
	LastActive int64
	ExpireAt   int64
}
