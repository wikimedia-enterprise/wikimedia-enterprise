package ksqldbarticles

import "time"

type Throttle struct {
	batchSize     int
	msgsPerSecond float32

	lastStarted time.Time
	msgsSoFar   int
}

func (t *Throttle) Start() {
	t.lastStarted = time.Now()
	t.msgsSoFar = 0
}

func (t *Throttle) increase() {
	t.msgsSoFar++
}

func (t *Throttle) throttleIfNeeded() {
	elapsed := time.Since(t.lastStarted)
	targetBatchTime := time.Duration(float32(time.Second) * float32(t.batchSize) / float32(t.msgsPerSecond))
	if targetBatchTime > elapsed {
		time.Sleep(targetBatchTime - elapsed)
	}
}

func (t *Throttle) Apply() {
	if t.msgsPerSecond <= 0 {
		// Disabled
		return
	}

	t.increase()
	if t.msgsSoFar >= t.batchSize {
		t.throttleIfNeeded()
		t.Start()
	}
}
