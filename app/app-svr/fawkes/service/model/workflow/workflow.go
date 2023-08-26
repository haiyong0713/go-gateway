package workflow

const (
	WorkflowOkStatus = 0
)

type Workflow struct {
	Title    string
	Name     string
	Operator string
	Params   map[string]interface{}
}
