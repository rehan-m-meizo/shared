package sagaflow

import (
	"context"
	"fmt"
	"shared/sagakit"
)

func StartSaga(ctx context.Context, s Saga) error {
	for i, st := range s.Steps {
		msg := map[string]any{
			"saga_id":    s.SagaID,
			"step_id":    st.ID,
			"step_index": i,
			"command":    st.Command,
			"payload":    st.Payload,
		}
		topic := fmt.Sprintf("saga.command.%s", st.Service)
		if err := sagakit.Publish(ctx, topic, msg); err != nil {
			return err
		}
	}
	return nil
}
