package model

import "time"

type CookieRemember struct {
	MemberId int
	Account  string
	Time     time.Time
}
