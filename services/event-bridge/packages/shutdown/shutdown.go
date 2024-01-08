package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Helper is an auxiliary struct which exposes the Wait method
// to be able to block execution until all goroutines are done.
type Helper struct {
	ctx     context.Context
	cancel  context.CancelFunc
	signals chan os.Signal
	wg      *sync.WaitGroup
}

func NewHelper(ctx context.Context) *Helper {
	h := new(Helper)
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.wg = new(sync.WaitGroup)
	h.signals = make(chan os.Signal, 1)
	signal.Notify(h.signals, os.Interrupt, syscall.SIGTERM)
	return h
}

// Ctx get helper context.
func (h *Helper) Ctx() context.Context {
	return h.ctx
}

// Kill sends an interrupt signal to the helper's signals channel.
func (h *Helper) Kill() {
	h.signals <- os.Interrupt
}

// WG returns the helper's wait group.
func (h *Helper) WG() *sync.WaitGroup {
	return h.wg
}

// Wait will listen for interrupt signals and lock the execution
// until the wait group is done.
func (h *Helper) Wait(p *kafka.Producer) {
	// Lock until an interrupt signal is received
	<-h.signals

	// Cancel the context and wait for all goroutines to finish
	h.cancel()
	h.wg.Wait()

	// Wait for kafka to flush all messages
	if p != nil {
		defer p.Close()

		for {
			if p.Flush(10000) == 0 {
				break
			}
		}
	}
}
