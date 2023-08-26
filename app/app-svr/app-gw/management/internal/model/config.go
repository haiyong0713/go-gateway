package model

type ConfigBuildItem struct {
	TreeID   int64         `json:"tree_id"`
	AppName  string        `json:"app_name"`
	Name     string        `json:"name"`
	Env      string        `json:"env"`
	Zone     string        `json:"zone"`
	Snapshot Snapshot      `json:"snapshot"`
	AppItem  ConfigAppItem `json:"app_item"`
}

type Snapshot struct {
	Files []ConfigMetadata `json:"files"`
}

type ConfigMetadata struct {
	ID       int64    `json:"id"`
	NewstID  int64    `json:"newst_id"`
	Name     string   `json:"name"`
	Envs     []string `json:"envs"`
	Zones    []string `json:"zones"`
	Operator string   `json:"operator"`
	IsDelete bool     `json:"is_delete"`
}

type ConfigAppItem struct {
	ID    int64  `json:"id"`
	Env   string `json:"env"`
	Zone  string `json:"zone"`
	Token string `json:"token"`
}
