package model

import "time"

type Bucket struct {
	Key      string
	Capacity int
	TTL      time.Duration
}
