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
			svc.StdoutTail = helpers.TruncateTail(t, 4096)
		}
		if t := helpers.StringFromAny(meta["stderr_tail"]); t != "" {
			svc.StderrTail = helpers.TruncateTail(t, 4096)
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

func applyCronJobResult(ctx context.Context, job AgentJob, ok bool) error {
	if job.Kind != constants.JobKindCronSync {
		return nil
	}
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
}

func ApplyProcessJobResult(ctx context.Context, job AgentJob, ok bool, meta map[string]any) error {
	switch job.Kind {
	case constants.JobKindServiceApply, constants.JobKindServiceRemove, constants.JobKindServiceControl:
		return applyServiceJobResult(ctx, job, ok, meta)
	case constants.JobKindCronSync:
		return applyCronJobResult(ctx, job, ok)
	default:
		return nil
	}
}
