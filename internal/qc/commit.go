package qc

// CommitOutcome summarizes a QC batch commit.
type CommitOutcome struct {
	PlanID       string   `json:"plan_id"`
	Committed    int      `json:"committed"`
	Failed       int      `json:"failed"`
	PendingRetry []string `json:"pending_retry"`
	Completed    bool     `json:"completed"`
}

// CommitBatch writes each device result of the batch. When a write fails
// midway, every remaining device is tracked for retry so the round is not
// silently lost.
func (s *ScheduleService) CommitBatch(plan *Plan, batch *Batch) (CommitOutcome, error) {
	committed := 0
	for i, item := range batch.Items {
		if err := s.batch.WriteResult(plan.ID, item.DeviceID, item.Result); err != nil {
			remaining := batch.RemainingIDs(i)
			if len(remaining) > 0 {
				_ = s.batch.MarkPending(plan.ID, remaining)
				plan.PendingRetry = append(plan.PendingRetry, remaining...)
				_ = s.Save(plan)
			}
			return CommitOutcome{
				PlanID:       plan.ID,
				Committed:    committed,
				Failed:       batch.Size() - committed,
				PendingRetry: remaining,
			}, err
		}
		committed++
	}
	_ = s.batch.MarkPending(plan.ID, nil)
	plan.PendingRetry = nil
	_ = s.Save(plan)
	return CommitOutcome{PlanID: plan.ID, Committed: committed, Completed: true}, nil
}

// RecoverPending reports which devices of a previously failed commit still
// need their results written.
func (s *ScheduleService) RecoverPending(plan *Plan) ([]string, error) {
	pending, ok := s.batch.Pending(plan.ID)
	if !ok {
		return nil, nil
	}
	missing := make([]string, 0, len(pending))
	for _, id := range pending {
		if !s.batch.ResultExists(plan.ID, id) {
			missing = append(missing, id)
		}
	}
	return missing, nil
}
