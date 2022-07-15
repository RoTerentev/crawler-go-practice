package internal

import (
	"log"
	"sync"
	"time"
)

type RateLimiter struct {
	counter int64
	limitCh chan struct{}
}

func NewLimiter(duration time.Duration, limit int64) *RateLimiter {
	rl := RateLimiter{
		counter: 0,
		limitCh: make(chan struct{}, limit),
	}
	ticker := time.NewTicker(duration)
	mu := sync.Mutex{}

	go func() {
		for _ = range rl.limitCh {
			mu.Lock()
			rl.counter++
			if rl.counter == limit {
				rl.counter = 0
				<-ticker.C
			}
			mu.Unlock()
		}

		log.Println("end limit ch")
	}()

	return &rl
}

func (limiter *RateLimiter) Add() chan<- struct{} {
	return limiter.limitCh
}
