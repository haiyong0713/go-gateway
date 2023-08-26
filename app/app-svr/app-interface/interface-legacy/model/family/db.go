package family

const (
	// family_relation.state
	RelStateUnbind = 0
	RelStateBind   = 1
	// family_log.operator
	OperatorUser = "user"
	// family_relation.timelock_state
	TlStateClose = 0
	TlStateOpen  = 1
)

type FamilyRelation struct {
	ID            int64 `json:"id"`
	ParentMid     int64 `json:"parent_mid"`
	ChildMid      int64 `json:"child_mid"`
	TimelockState int64 `json:"timelock_state"`
	DailyDuration int64 `json:"daily_duration"`
}

type FamilyLog struct {
	ID       int64  `json:"id"`
	Mid      int64  `json:"mid"`
	Operator string `json:"operator"`
	Content  string `json:"content"`
}
