package sagaflow

import (
	"context"
	"encoding/json"
	"fmt"
	"shared/sagakit/db"
	"time"
)

// SagaState represents the current state of a saga execution
type SagaState string

const (
	StateCreated      SagaState = "CREATED"
	StateInProgress   SagaState = "IN_PROGRESS"
	StateCompleted    SagaState = "COMPLETED"
	StateFailed       SagaState = "FAILED"
	StateCompensating SagaState = "COMPENSATING"
	StateCompensated  SagaState = "COMPENSATED"
)

// StepState represents the state of individual saga steps
type StepState string

const (
	StepPending      StepState = "PENDING"
	StepInProgress   StepState = "IN_PROGRESS"
	StepCompleted    StepState = "COMPLETED"
	StepFailed       StepState = "FAILED"
	StepCompensating StepState = "COMPENSATING"
	StepCompensated  StepState = "COMPENSATED"
)

// SagaExecution tracks the execution of a saga
type SagaExecution struct {
	SagaID       string                 `json:"saga_id"`
	SagaName     string                 `json:"saga_name"`
	State        SagaState              `json:"state"`
	CurrentStep  int                    `json:"current_step"`
	Steps        []StepExecution        `json:"steps"`
	Context      map[string]interface{} `json:"context"` // Shared data between steps
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// StepExecution tracks individual step execution
type StepExecution struct {
	StepID       string                 `json:"step_id"`
	StepIndex    int                    `json:"step_index"`
	State        StepState              `json:"state"`
	Service      string                 `json:"service"`
	Command      string                 `json:"command"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Attempts     int                    `json:"attempts"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// StateStore handles saga state persistence
type StateStore interface {
	SaveExecution(ctx context.Context, tx db.Tx, exec *SagaExecution) error
	GetExecution(ctx context.Context, tx db.Tx, sagaID string) (*SagaExecution, error)
	UpdateStepState(ctx context.Context, tx db.Tx, sagaID string, stepIndex int, state StepState, output map[string]interface{}, errMsg string) error
	UpdateSagaState(ctx context.Context, tx db.Tx, sagaID string, state SagaState, currentStep int, errMsg string) error
}

// PostgresStateStore implements StateStore for PostgreSQL
type PostgresStateStore struct{}

func NewPostgresStateStore() *PostgresStateStore {
	return &PostgresStateStore{}
}

// InitSchema creates the necessary tables
func (s *PostgresStateStore) InitSchema(ctx context.Context, tx db.Tx) error {
	schema := `
		CREATE TABLE IF NOT EXISTS saga_executions (
			saga_id TEXT PRIMARY KEY,
			saga_name TEXT NOT NULL,
			state TEXT NOT NULL,
			current_step INT NOT NULL DEFAULT 0,
			context JSONB,
			error_message TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS saga_step_executions (
			saga_id TEXT NOT NULL,
			step_index INT NOT NULL,
			step_id TEXT NOT NULL,
			state TEXT NOT NULL,
			service TEXT NOT NULL,
			command TEXT NOT NULL,
			input JSONB,
			output JSONB,
			error_message TEXT,
			attempts INT NOT NULL DEFAULT 0,
			started_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			PRIMARY KEY (saga_id, step_index)
		);

		CREATE INDEX IF NOT EXISTS idx_saga_executions_state ON saga_executions(state);
		CREATE INDEX IF NOT EXISTS idx_saga_step_executions_saga_id ON saga_step_executions(saga_id);
	`
	return tx.Exec(ctx, schema)
}

func (s *PostgresStateStore) SaveExecution(ctx context.Context, tx db.Tx, exec *SagaExecution) error {
	contextJSON, err := json.Marshal(exec.Context)
	if err != nil {
		return err
	}

	// Save saga execution
	err = tx.Exec(ctx, `
		INSERT INTO saga_executions (saga_id, saga_name, state, current_step, context, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (saga_id) DO UPDATE SET
			state = EXCLUDED.state,
			current_step = EXCLUDED.current_step,
			context = EXCLUDED.context,
			error_message = EXCLUDED.error_message,
			updated_at = EXCLUDED.updated_at
	`, exec.SagaID, exec.SagaName, exec.State, exec.CurrentStep, contextJSON, exec.ErrorMessage, exec.CreatedAt, exec.UpdatedAt)
	if err != nil {
		return err
	}

	// Save step executions
	for _, step := range exec.Steps {
		inputJSON, _ := json.Marshal(step.Input)
		outputJSON, _ := json.Marshal(step.Output)

		err = tx.Exec(ctx, `
			INSERT INTO saga_step_executions (saga_id, step_index, step_id, state, service, command, input, output, error_message, attempts, started_at, completed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (saga_id, step_index) DO UPDATE SET
				state = EXCLUDED.state,
				output = EXCLUDED.output,
				error_message = EXCLUDED.error_message,
				attempts = EXCLUDED.attempts,
				started_at = EXCLUDED.started_at,
				completed_at = EXCLUDED.completed_at
		`, exec.SagaID, step.StepIndex, step.StepID, step.State, step.Service, step.Command, inputJSON, outputJSON, step.ErrorMessage, step.Attempts, step.StartedAt, step.CompletedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStateStore) GetExecution(ctx context.Context, tx db.Tx, sagaID string) (*SagaExecution, error) {
	rows, err := tx.Query(ctx, `
		SELECT saga_id, saga_name, state, current_step, context, error_message, created_at, updated_at
		FROM saga_executions
		WHERE saga_id = $1
	`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("saga not found: %s", sagaID)
	}

	var exec SagaExecution
	var contextJSON []byte
	err = rows.Scan(&exec.SagaID, &exec.SagaName, &exec.State, &exec.CurrentStep, &contextJSON, &exec.ErrorMessage, &exec.CreatedAt, &exec.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if len(contextJSON) > 0 {
		json.Unmarshal(contextJSON, &exec.Context)
	}

	// Load steps
	stepRows, err := tx.Query(ctx, `
		SELECT step_id, step_index, state, service, command, input, output, error_message, attempts, started_at, completed_at
		FROM saga_step_executions
		WHERE saga_id = $1
		ORDER BY step_index
	`, sagaID)
	if err != nil {
		return nil, err
	}
	defer stepRows.Close()

	for stepRows.Next() {
		var step StepExecution
		var inputJSON, outputJSON []byte
		err = stepRows.Scan(&step.StepID, &step.StepIndex, &step.State, &step.Service, &step.Command, &inputJSON, &outputJSON, &step.ErrorMessage, &step.Attempts, &step.StartedAt, &step.CompletedAt)
		if err != nil {
			return nil, err
		}

		if len(inputJSON) > 0 {
			json.Unmarshal(inputJSON, &step.Input)
		}
		if len(outputJSON) > 0 {
			json.Unmarshal(outputJSON, &step.Output)
		}

		exec.Steps = append(exec.Steps, step)
	}

	return &exec, nil
}

func (s *PostgresStateStore) UpdateStepState(ctx context.Context, tx db.Tx, sagaID string, stepIndex int, state StepState, output map[string]interface{}, errMsg string) error {
	outputJSON, _ := json.Marshal(output)

	var completedAt *time.Time
	if state == StepCompleted || state == StepFailed || state == StepCompensated {
		now := time.Now()
		completedAt = &now
	}

	return tx.Exec(ctx, `
		UPDATE saga_step_executions
		SET state = $1, output = $2, error_message = $3, completed_at = $4, attempts = attempts + 1
		WHERE saga_id = $5 AND step_index = $6
	`, state, outputJSON, errMsg, completedAt, sagaID, stepIndex)
}

func (s *PostgresStateStore) UpdateSagaState(ctx context.Context, tx db.Tx, sagaID string, state SagaState, currentStep int, errMsg string) error {
	return tx.Exec(ctx, `
		UPDATE saga_executions
		SET state = $1, current_step = $2, error_message = $3, updated_at = now()
		WHERE saga_id = $4
	`, state, currentStep, errMsg, sagaID)
}
