package models

type RateLimit struct {
	IP          string
	WindowStart int64
	Count       int
	UpdatedAt   int64
}
