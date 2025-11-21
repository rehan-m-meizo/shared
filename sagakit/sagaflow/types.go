
package sagaflow

type Step struct {
    ID string
    Service string
    Command string
    Payload map[string]any
    Compensate string
    MaxRetries int
}
type Saga struct {
    SagaID string
    Name string
    Steps []Step
}
