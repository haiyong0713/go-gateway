package acg

type UserTaskState struct {
	Task       []*TaskState `json:"task"`
	FinishTask int          `json:"finish_task"`
	Money      int          `json:"money"`
}

type TaskState struct {
	Finish bool  `json:"finish"`
	Count  int   `json:"count"`
	Score  int64 `json:"score"`
}
