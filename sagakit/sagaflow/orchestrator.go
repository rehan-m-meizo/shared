package sagaflow

import (
	"context"
	"encoding/json"
	"fmt"
	"shared/pkgs/uuids"
	"shared/sagakit"
	"shared/sagakit/db"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Orchestrator manages saga execution
type Orchestrator struct {
	DB         db.DB
	StateStore StateStore
	Router     *message.Router
}

func NewOrchestrator(database db.DB, stateStore StateStore) *Orchestrator {
	return &Orchestrator{
		DB:         database,
		StateStore: stateStore,
	}
}

// StartSaga initiates a new saga execution
func (o *Orchestrator) StartSaga(ctx context.Context, sagaDef Saga, initialContext map[string]interface{}) (string, error) {
	sagaID := uuids.NewUUID()

	// Create saga execution
	exec := &SagaExecution{
		SagaID:      sagaID,
		SagaName:    sagaDef.Name,
		State:       StateCreated,
		CurrentStep: 0,
		Context:     initialContext,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Initialize step executions
	for i, step := range sagaDef.Steps {
		// Merge saga context with step payload
		stepInput := make(map[string]interface{})
		for k, v := range initialContext {
			stepInput[k] = v
		}
		for k, v := range step.Payload {
			stepInput[k] = v
		}

		exec.Steps = append(exec.Steps, StepExecution{
			StepID:    step.ID,
			StepIndex: i,
			State:     StepPending,
			Service:   step.Service,
			Command:   step.Command,
			Input:     stepInput,
			Attempts:  0,
		})
	}

	// Save initial state
	err := sagakit.RunInTx(ctx, o.DB, sagakit.GetGlobalStore(), func(uow sagakit.UnitOfWork) error {
		if err := o.StateStore.SaveExecution(ctx, uow.Tx(), exec); err != nil {
			return err
		}

		// Update state to in progress
		exec.State = StateInProgress
		if err := o.StateStore.UpdateSagaState(ctx, uow.Tx(), sagaID, StateInProgress, 0, ""); err != nil {
			return err
		}

		// Send first step command
		return o.sendStepCommand(uow, exec, 0)
	})

	if err != nil {
		return "", err
	}

	return sagaID, nil
}

// HandleStepSuccess processes successful step completion
func (o *Orchestrator) HandleStepSuccess(ctx context.Context, sagaID string, stepIndex int, output map[string]interface{}) error {
	return sagakit.RunInTx(ctx, o.DB, sagakit.GetGlobalStore(), func(uow sagakit.UnitOfWork) error {
		exec, err := o.StateStore.GetExecution(ctx, uow.Tx(), sagaID)
		if err != nil {
			return err
		}

		// Update step state
		if err := o.StateStore.UpdateStepState(ctx, uow.Tx(), sagaID, stepIndex, StepCompleted, output, ""); err != nil {
			return err
		}

		// Merge output into saga context
		for k, v := range output {
			exec.Context[k] = v
		}

		// Check if this was the last step
		if stepIndex == len(exec.Steps)-1 {
			// Saga completed successfully
			return o.StateStore.UpdateSagaState(ctx, uow.Tx(), sagaID, StateCompleted, stepIndex, "")
		}

		// Move to next step
		nextStepIndex := stepIndex + 1
		if err := o.StateStore.UpdateSagaState(ctx, uow.Tx(), sagaID, StateInProgress, nextStepIndex, ""); err != nil {
			return err
		}

		// Update step input with context
		exec.Steps[nextStepIndex].Input = exec.Context

		return o.sendStepCommand(uow, exec, nextStepIndex)
	})
}

// HandleStepFailure processes step failure
func (o *Orchestrator) HandleStepFailure(ctx context.Context, sagaID string, stepIndex int, errMsg string) error {
	return sagakit.RunInTx(ctx, o.DB, sagakit.GetGlobalStore(), func(uow sagakit.UnitOfWork) error {
		exec, err := o.StateStore.GetExecution(ctx, uow.Tx(), sagaID)
		if err != nil {
			return err
		}

		step := exec.Steps[stepIndex]

		// Check if we should retry
		if step.Attempts < exec.Steps[stepIndex].Attempts { // This would come from saga definition
			// Retry logic would go here
			return o.sendStepCommand(uow, exec, stepIndex)
		}

		// Mark step as failed
		if err := o.StateStore.UpdateStepState(ctx, uow.Tx(), sagaID, stepIndex, StepFailed, nil, errMsg); err != nil {
			return err
		}

		// Update saga state to compensating
		if err := o.StateStore.UpdateSagaState(ctx, uow.Tx(), sagaID, StateCompensating, stepIndex, errMsg); err != nil {
			return err
		}

		// Start compensation
		return o.startCompensation(uow, exec, stepIndex)
	})
}

// startCompensation initiates the compensation process
func (o *Orchestrator) startCompensation(uow sagakit.UnitOfWork, exec *SagaExecution, failedIndex int) error {
	// Compensate in reverse order
	for i := failedIndex - 1; i >= 0; i-- {
		step := exec.Steps[i]
		if step.State != StepCompleted {
			continue // Only compensate completed steps
		}

		msg := map[string]interface{}{
			"saga_id":    exec.SagaID,
			"step_id":    step.StepID,
			"step_index": i,
			"input":      step.Input,
			"output":     step.Output,
		}

		topic := fmt.Sprintf("saga.compensate.%s", step.Service)
		if err := uow.Publish(topic, msg, map[string]string{
			"saga_id":    exec.SagaID,
			"step_index": fmt.Sprintf("%d", i),
		}); err != nil {
			return err
		}

		// Update step state to compensating
		if err := o.StateStore.UpdateStepState(context.Background(), uow.Tx(), exec.SagaID, i, StepCompensating, nil, ""); err != nil {
			return err
		}
	}

	return nil
}

// HandleCompensationSuccess processes successful compensation
func (o *Orchestrator) HandleCompensationSuccess(ctx context.Context, sagaID string, stepIndex int) error {
	return sagakit.RunInTx(ctx, o.DB, sagakit.GetGlobalStore(), func(uow sagakit.UnitOfWork) error {
		exec, err := o.StateStore.GetExecution(ctx, uow.Tx(), sagaID)
		if err != nil {
			return err
		}

		// Update step state
		if err := o.StateStore.UpdateStepState(ctx, uow.Tx(), sagaID, stepIndex, StepCompensated, nil, ""); err != nil {
			return err
		}

		// Check if all compensations are done
		allCompensated := true
		for i, step := range exec.Steps {
			if i >= exec.CurrentStep {
				break
			}
			if step.State == StepCompleted && exec.Steps[i].State != StepCompensated {
				allCompensated = false
				break
			}
		}

		if allCompensated {
			return o.StateStore.UpdateSagaState(ctx, uow.Tx(), sagaID, StateCompensated, exec.CurrentStep, exec.ErrorMessage)
		}

		return nil
	})
}

// sendStepCommand publishes a command for a saga step
func (o *Orchestrator) sendStepCommand(uow sagakit.UnitOfWork, exec *SagaExecution, stepIndex int) error {
	step := exec.Steps[stepIndex]

	msg := map[string]interface{}{
		"saga_id":    exec.SagaID,
		"step_id":    step.StepID,
		"step_index": stepIndex,
		"command":    step.Command,
		"input":      step.Input,
	}

	topic := fmt.Sprintf("saga.command.%s", step.Service)
	return uow.Publish(topic, msg, map[string]string{
		"saga_id":    exec.SagaID,
		"step_index": fmt.Sprintf("%d", stepIndex),
	})
}

// SetupEventHandlers subscribes to saga events
func (o *Orchestrator) SetupEventHandlers(ctx context.Context, subscriber message.Subscriber) error {
	router, err := message.NewRouter(message.RouterConfig{}, sagakit.GetLogger())
	if err != nil {
		return err
	}

	// Handle step success events
	router.AddConsumerHandler(
		"saga_step_success",
		"saga.event.step.success",
		subscriber,
		func(msg *message.Message) error {
			var event struct {
				SagaID    string                 `json:"saga_id"`
				StepIndex int                    `json:"step_index"`
				Output    map[string]interface{} `json:"output"`
			}
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				return err
			}
			return o.HandleStepSuccess(ctx, event.SagaID, event.StepIndex, event.Output)
		},
	)

	// Handle step failure events
	router.AddConsumerHandler(
		"saga_step_failure",
		"saga.event.step.failure",
		subscriber,
		func(msg *message.Message) error {
			var event struct {
				SagaID    string `json:"saga_id"`
				StepIndex int    `json:"step_index"`
				Error     string `json:"error"`
			}
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				return err
			}
			return o.HandleStepFailure(ctx, event.SagaID, event.StepIndex, event.Error)
		},
	)

	// Handle compensation success events
	router.AddConsumerHandler(
		"saga_compensation_success",
		"saga.event.compensation.success",
		subscriber,
		func(msg *message.Message) error {
			var event struct {
				SagaID    string `json:"saga_id"`
				StepIndex int    `json:"step_index"`
			}
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				return err
			}
			return o.HandleCompensationSuccess(ctx, event.SagaID, event.StepIndex)
		},
	)

	o.Router = router
	go func() {
		if err := router.Run(ctx); err != nil {
			fmt.Printf("Router error: %v\n", err)
		}
	}()

	<-router.Running()
	return nil
}
