package models

type BlockedIP struct {
	IP        string
	BlockedAt int64
	ExpireAt  int64
}
