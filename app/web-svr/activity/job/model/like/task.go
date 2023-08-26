package like

type TaskUserLog struct {
	ID         int64 `json:"id"`
	Mid        int64 `json:"mid"`
	BusinessID int64 `json:"business_id"`
	TaskID     int64 `json:"task_id"`
	ForeignID  int64 `json:"foreign_id"`
}
