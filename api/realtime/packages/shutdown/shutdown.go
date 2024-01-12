package shutdown

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wikimedia-enterprise/general/log"
)

// Helper is an auxiliary struct which exposes the Wait method
// to be able to block execution until API server shuts down gracefully.
type Helper struct {
	ctx     context.Context
	cancel  context.CancelFunc
	signals chan os.Signal
}

// NewHelper creates Helper instance. It takes a context as an argument.
func NewHelper(ctx context.Context) *Helper {
	h := new(Helper)
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.signals = make(chan os.Signal, 1)
	signal.Notify(h.signals, os.Interrupt, syscall.SIGTERM)
	return h
}

// Ctx returns helper context.
func (h *Helper) Ctx() context.Context {
	return h.ctx
}

// Kill sends an interrupt signal to the helper's signals channel. Used in unit testing.
func (h *Helper) Kill() {
	h.signals <- os.Interrupt
}

// Wait takes an interface with close() as an argument and listens for an interrupt signal.
// When an interrupt is received on the channel, it cancels the helper context and shuts the server down.
func (h *Helper) Wait(srv io.Closer) error {
	<-h.signals
	h.cancel()
	log.Info("received interrupt signal, shutting down...")

	time.Sleep(time.Second * 5)

	if err := srv.Close(); err != nil {
		log.Error(err, log.Tip("problem in closing server"))
		return err
	}

	log.Info("server shut down")

	return nil
}
