package models

import (
	"context"

	"api/app/constants"
	"api/app/helpers"
)

func applyServiceJobResult(ctx context.Context, job AgentJob, ok bool, meta map[string]any) error {
	uid, _ := job.Payload["uid"].(string)
	name, _ := job.Payload["name"].(string)

	var svc ServerService
	var err error
	switch {
	case uid != "":
		svc, err = (ServerService{}).FindByPublicID(ctx, uid, job.ServerID)
	case name != "":
		svc, err = (ServerService{}).FindByName(ctx, job.ServerID, name)
	default:
		return nil
	}
	if err != nil {
		return nil
	}

	msg := job.Message
	if meta != nil {
		if s := helpers.StringFromAny(meta["status"]); s != "" && constants.IsValidServiceStatus(s) {
			svc.Status = s
		}
		if t := helpers.StringFromAny(meta["stdout_tail"]); t != "" {
			svc.StdoutTail = helpers.TruncateTail(t, 16384)
		}
		if t := helpers.StringFromAny(meta["stderr_tail"]); t != "" {
			svc.StderrTail = helpers.TruncateTail(t, 16384)
		}
	}

	switch job.Kind {
	case constants.JobKindServiceRemove:
		if ok {
			return svc.Delete(ctx)
		}
		svc.Status = constants.ServiceStatusFailed
		svc.Message = msg
		return svc.Save(ctx)

	case constants.JobKindServiceApply, constants.JobKindServiceControl:
		if !ok {
			svc.Status = constants.ServiceStatusFailed
			svc.Message = msg
			return svc.Save(ctx)
		}
		if svc.Status == constants.ServiceStatusPending || svc.Status == "" {
			svc.Status = constants.ServiceStatusUnknown
		}
		if !svc.Enabled && job.Kind == constants.JobKindServiceApply {
			svc.Status = constants.ServiceStatusStopped
		}
		svc.Message = msg
		return svc.Save(ctx)
	}
	return nil
}

func applyCronJobResult(ctx context.Context, job AgentJob, ok bool, meta map[string]any) error {
	switch job.Kind {
	case constants.JobKindCronSync:
		crons, err := (ServerCron{}).ListByServer(ctx, job.ServerID)
		if err != nil {
			return err
		}
		status := constants.CronStatusSynced
		if !ok {
			status = constants.CronStatusFailed
		}
		for i := range crons {
			crons[i].Status = status
			crons[i].Message = job.Message
			if err := crons[i].Save(ctx); err != nil {
				return err
			}
		}
		return nil

	case constants.JobKindCronLogs:
		uid, _ := job.Payload["uid"].(string)
		if uid == "" {
			return nil
		}
		cron, err := (ServerCron{}).FindByPublicID(ctx, uid, job.ServerID)
		if err != nil {
			return nil
		}
		if meta != nil {
			if t := helpers.StringFromAny(meta["log_tail"]); t != "" {
				cron.LogTail = helpers.TruncateTail(t, 4096)
			}
		}
		if !ok {
			cron.Message = job.Message
		} else if cron.LogTail == "" {
			cron.LogTail = "(empty)"
		}
		return cron.Save(ctx)
	}
	return nil
}

func applyFirewallJobResult(ctx context.Context, job AgentJob, ok bool, _ map[string]any) error {
	if job.Kind != constants.JobKindFirewallSync {
		return nil
	}
	rules, err := (ServerFirewallRule{}).ListByServer(ctx, job.ServerID)
	if err != nil {
		return err
	}
	status := constants.FirewallStatusSynced
	if !ok {
		status = constants.FirewallStatusFailed
	}
	for i := range rules {
		rules[i].Status = status
		rules[i].Message = job.Message
		if err := rules[i].Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func ApplyProcessJobResult(ctx context.Context, job AgentJob, ok bool, meta map[string]any) error {
	switch job.Kind {
	case constants.JobKindServiceApply, constants.JobKindServiceRemove, constants.JobKindServiceControl:
		return applyServiceJobResult(ctx, job, ok, meta)
	case constants.JobKindCronSync, constants.JobKindCronLogs:
		return applyCronJobResult(ctx, job, ok, meta)
	case constants.JobKindFirewallSync:
		return applyFirewallJobResult(ctx, job, ok, meta)
	default:
		return nil
	}
}
