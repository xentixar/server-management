// Package executor runs allowlisted job kinds. Add handlers here as capabilities grow.
package executor

import (
	"context"
	"fmt"

	"agent/internal/protocol"
)

type Handler func(ctx context.Context, job protocol.Job) (string, error)

type Executor struct {
	handlers map[string]Handler
}

func New(handlers map[string]Handler) *Executor {
	if handlers == nil {
		handlers = map[string]Handler{}
	}
	return &Executor{handlers: handlers}
}

// Default registers built-in kinds. Unknown kinds fail closed.
func Default() *Executor {
	return New(map[string]Handler{
		"ping": handlePing,
	})
}

func (e *Executor) Run(ctx context.Context, job protocol.Job) protocol.JobResultRequest {
	h, ok := e.handlers[job.Kind]
	if !ok {
		return protocol.JobResultRequest{
			OK:       false,
			Message:  fmt.Sprintf("unsupported job kind %q", job.Kind),
			ExitCode: 1,
		}
	}

	msg, err := h(ctx, job)
	if err != nil {
		return protocol.JobResultRequest{
			OK:       false,
			Message:  err.Error(),
			ExitCode: 1,
		}
	}
	return protocol.JobResultRequest{OK: true, Message: msg}
}

func handlePing(_ context.Context, _ protocol.Job) (string, error) {
	return "pong", nil
}
