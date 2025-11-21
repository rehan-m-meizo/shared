package sagaflow

import (
	"context"
	"fmt"
	"shared/sagakit"
	"time"
)

func RetryStep(ctx context.Context, s Saga, stepIndex int, attempt int) error {
	st := s.Steps[stepIndex]
	if attempt > st.MaxRetries {
		return fmt.Errorf("max retries exceeded")
	}
	time.Sleep(time.Duration(attempt) * time.Second)
	msg := map[string]any{
		"saga_id":    s.SagaID,
		"step_id":    st.ID,
		"step_index": stepIndex,
		"command":    st.Command,
		"payload":    st.Payload,
	}
	topic := fmt.Sprintf("saga.command.%s", st.Service)
	return sagakit.Publish(ctx, topic, msg)
}
