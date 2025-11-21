package sagaflow

import (
	"context"
	"fmt"
	"shared/sagakit"
)

func RunCompensation(ctx context.Context, s Saga, failedIndex int) error {
	for i := failedIndex - 1; i >= 0; i-- {
		st := s.Steps[i]
		if st.Compensate == "" {
			continue
		}
		msg := map[string]any{
			"saga_id":    s.SagaID,
			"step_id":    st.ID,
			"step_index": i,
			"command":    st.Compensate,
			"payload":    st.Payload,
		}
		topic := fmt.Sprintf("saga.compensate.%s", st.Service)
		if err := sagakit.Publish(ctx, topic, msg); err != nil {
			return err
		}
	}
	return nil
}
