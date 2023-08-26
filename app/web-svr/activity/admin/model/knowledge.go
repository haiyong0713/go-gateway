package model

type KnowConfig struct {
	ID            int64  `json:"id"`
	ConfigDetails string `json:"config_details"`
}

// TableName KnowConfig .
func (KnowConfig) TableName() string {
	return "act_knowledge_config"
}

// KnowConfigInfo .
type KnowConfigInfo struct {
	ID            int64             `json:"id"`
	ConfigDetails *KnowConfigDetail `json:"config_details"`
}

type KnowTask struct {
	TaskName   string  `json:"task_name"`
	TaskColumn string  `json:"task_column"`
	TaskFinish int64   `json:"task_finish"`
	Priority   float64 `json:"priority"`
	Image      string  `json:"image"`
	ImageGrey  string  `json:"image_grey"`
	IsFinish   bool    `json:"is_finish"`
}

type LevelTask struct {
	Priority          float64     `json:"priority"`
	ParentName        string      `json:"parent_name"`
	ParentDescription string      `json:"parent_description"`
	Tasks             []*KnowTask `json:"tasks"`
}

type KnowConfigDetail struct {
	Name          string                `json:"name"`
	Table         string                `json:"table"`
	TableCount    int64                 `json:"table_count"`
	TotalBadge    float64               `json:"total_badge"`
	ShareFields   string                `json:"share_fields"`
	TotalProgress float64               `json:"total_progress"`
	LevelTask     map[string]*LevelTask `json:"level_task"`
}

type ParamKnowledge struct {
	ConfigId   int64   `form:"config_id" validate:"required"`
	TaskName   string  `form:"task_name" validate:"required"`
	UpdateMids []int64 `form:"update_mids,split" validate:"min=1,max=500,dive,min=1"`
	IsBack     bool    `form:"is_back"`
}
