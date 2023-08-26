package fawkes

// fawkes consts.
const (
	StatusUpFaild   = -2
	StatusSendFaild = -1
	StatusQueuing   = 1
	StatusWaitSend  = 2
	StatusUpSuccess = 3
)

// LaserMsg struct.
type LaserMsg struct {
	Date   string `json:"date"`
	TaskID string `json:"taskid"`
}
