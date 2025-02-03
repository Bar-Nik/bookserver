package domain

import "time"

type Session struct {
	ID        int
	UserID    int
	Token     string
	IP        string
	UserAgent string
	CreatedAt time.Time
}
