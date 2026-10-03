// Package loop is the agent runtime: heartbeat + job poll until context cancel.
package loop

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"agent/internal/client"
	"agent/internal/executor"
	"agent/internal/metrics"
	"agent/internal/protocol"
)

type ControlPlane interface {
	Heartbeat(ctx context.Context, req protocol.HeartbeatRequest) (protocol.HeartbeatResponse, error)
	NextJob(ctx context.Context) (*protocol.Job, error)
	ReportJob(ctx context.Context, jobID string, result protocol.JobResultRequest) error
}

type Options struct {
	HeartbeatInterval time.Duration
	PollInterval      time.Duration
	Hostname          string
}

type Loop struct {
	api     ControlPlane
	exec    *executor.Executor
	opts    Options
	log     *slog.Logger
	metrics *metrics.Collector
}

func New(api ControlPlane, exec *executor.Executor, opts Options, log *slog.Logger) *Loop {
	if log == nil {
		log = slog.Default()
	}
	if opts.Hostname == "" {
		opts.Hostname, _ = os.Hostname()
	}
	if exec == nil {
		exec = executor.Default()
	}
	return &Loop{api: api, exec: exec, opts: opts, log: log, metrics: &metrics.Collector{}}
}

// Run blocks until ctx is cancelled. Transient API errors are logged and retried.
func (l *Loop) Run(ctx context.Context) error {
	l.log.Info("agent loop starting",
		"protocol", protocol.Version,
		"version", protocol.AgentVersion,
		"heartbeat", l.opts.HeartbeatInterval,
		"poll", l.opts.PollInterval,
	)

	// Immediate heartbeat so status flips pending→online without waiting a full tick.
	l.heartbeat(ctx)

	heartbeat := time.NewTicker(l.opts.HeartbeatInterval)
	poll := time.NewTicker(l.opts.PollInterval)
	defer heartbeat.Stop()
	defer poll.Stop()

	for {
		select {
		case <-ctx.Done():
			l.log.Info("agent loop stopping", "reason", ctx.Err())
			return ctx.Err()
		case <-heartbeat.C:
			l.heartbeat(ctx)
		case <-poll.C:
			l.pollOnce(ctx)
		}
	}
}

func (l *Loop) heartbeat(ctx context.Context) {
	resp, err := l.api.Heartbeat(ctx, protocol.HeartbeatRequest{
		ProtocolVersion: protocol.Version,
		AgentVersion:    protocol.AgentVersion,
		Hostname:        l.opts.Hostname,
		Metrics:         l.metrics.Sample(),
	})
	if err != nil {
		if errors.Is(err, client.ErrUnauthorized) {
			l.log.Error("heartbeat unauthorized; check SERVER_TOKEN")
			return
		}
		l.log.Warn("heartbeat failed", "err", err)
		return
	}
	if resp.ProtocolVersion != 0 && resp.ProtocolVersion != protocol.Version {
		l.log.Warn("protocol mismatch",
			"agent", protocol.Version,
			"server", resp.ProtocolVersion,
		)
	}
	l.log.Debug("heartbeat ok")
}

func (l *Loop) pollOnce(ctx context.Context) {
	job, err := l.api.NextJob(ctx)
	if err != nil {
		if errors.Is(err, client.ErrUnauthorized) {
			l.log.Error("job poll unauthorized; check SERVER_TOKEN")
			return
		}
		l.log.Warn("job poll failed", "err", err)
		return
	}
	if job == nil {
		return
	}

	l.log.Info("job claimed", "id", job.ID, "kind", job.Kind)
	result := l.exec.Run(ctx, *job)
	if err := l.api.ReportJob(ctx, job.ID, result); err != nil {
		l.log.Error("job result report failed", "id", job.ID, "err", err)
		return
	}
	if result.OK {
		l.log.Info("job done", "id", job.ID, "kind", job.Kind)
	} else {
		l.log.Error("job failed", "id", job.ID, "kind", job.Kind, "msg", result.Message)
	}
}
