package helpers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"code.cestus.io/tools/fabricator/pkg/fabricator"
)

// WithCancelOnSignal returns a context that will get cancelled whenever one of
// the specified signals is caught.
func WithCancelOnSignal(ctx context.Context, io fabricator.IOStreams, signals ...os.Signal) (context.Context, func()) {
	var once sync.Once
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal)

	signal.Notify(ch, signals...)

	go func() {
		if sig := <-ch; sig != nil {
			fmt.Fprintf(io.Out, "Context received signal: %s\n", sig)
		}

		cancel()
	}()

	return ctx, func() {
		once.Do(func() {
			signal.Reset(signals...)
			close(ch)
		})
		<-ctx.Done()
	}
}
