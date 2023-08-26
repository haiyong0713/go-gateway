package saga

const (
	//HookCommentTypeMR ...
	HookCommentTypeMR = "MergeRequest"
)

// HookPush event module
type HookPush struct {
	ObjectKind        string      `json:"object_kind"`
	Before            string      `json:"before"`
	After             string      `json:"after"`
	Ref               string      `json:"ref"` // 分支 refs/heads/xxx
	CheckoutSHA       string      `json:"checkout_sha"`
	UserID            int64       `json:"user_id"`
	UserName          string      `json:"user_name"`
	UserUserName      string      `json:"user_username"`
	UserEmail         string      `json:"user_email"`
	UserAvatar        string      `json:"user_avatar"`
	ProjectID         int64       `json:"project_id"`
	Project           *Project    `json:"project"`
	Repository        *Repository `json:"repository"`
	Commits           []*Commit   `json:"commits"`
	TotalCommitsCount int64       `json:"total_commits_count"`
}

// HookMR def
type HookMR struct {
	ObjectKind       string        `json:"object_kind"`
	Project          *Project      `json:"project"`
	User             *User         `json:"user"`
	ObjectAttributes *MergeRequest `json:"object_attributes"`
	Assignee         *User         `json:"assignee"`
}

// MergeRequest struct
type MergeRequest struct {
	ID              int64    `json:"id"`
	TargetBranch    string   `json:"target_branch"`
	SourceBranch    string   `json:"source_branch"`
	SourceProjectID int64    `json:"source_project_id"`
	AuthorID        int64    `json:"author_id"`
	AssigneeID      int64    `json:"assignee_id"`
	Title           string   `json:"title"`
	CreateAt        string   `json:"created_at"`
	UpdateAt        string   `json:"updated_at"`
	STCommits       int64    `json:"st_commits"`
	STDiffs         int64    `json:"st_diffs"`
	MilestoneID     int64    `json:"milestone_id"`
	State           string   `json:"state"`
	MergeStatus     string   `json:"merge_status"`
	TargetProjectID int64    `json:"target_project_id"`
	IID             int64    `json:"iid"`
	Description     string   `json:"description"`
	Source          *Project `json:"source"`
	Target          *Project `json:"target"`
	LastCommit      *Commit  `json:"last_commit"`
	WorkInProgress  bool     `json:"work_in_progress"`
	URL             string   `json:"url"`
	Action          string   `json:"action"` // "open","update","close"
	Sha             string   `json:"sha"`
}

// Project def
type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	AvatarURL         string `json:"avatar_url"`
	GitSSHURL         string `json:"git_ssh_url"`
	GitHTTPURL        string `json:"git_http_url"`
	Namespace         string `json:"namespace"`
	VisibilityLevel   int64  `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
	Homepage          string `json:"homepage"`
	URL               string `json:"url"`
	SSHURL            string `json:"ssh_url"`
	HTTPURL           string `json:"http_url"`
}

// Repository def
type Repository struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Description     string `json:"description"`
	Homepage        string `json:"homepage"`
	GitHTTPURL      string `json:"git_http_url"`
	GitSSHURL       string `json:"git_ssh_url"`
	VisibilityLevel int64  `json:"visibility_level"`
}

// Commit def
type Commit struct {
	ID        string   `json:"id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    *Author  `json:"author"`
	Added     []string `json:"added"`
	Modified  []string `json:"modified"`
	Removed   []string `json:"removed"`
}

// Author def
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User def
type User struct {
	Name      string `json:"name"`
	UserName  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

// HookComment struct
type HookComment struct {
	ObjectKind       string        `json:"object_kind"`
	User             *User         `json:"user"`
	ProjectID        int64         `json:"project_id"`
	Project          *Project      `json:"project"`
	Repository       *Repository   `json:"repository"`
	ObjectAttributes *Comment      `json:"object_attributes"`
	MergeRequest     *MergeRequest `json:"merge_request"`
	Commit           *Commit       `json:"commit"`
}

// Comment struct
type Comment struct {
	ID           int64  `json:"id"`
	Note         string `json:"note"`
	NoteableType string `json:"noteable_type"`
	AuthorID     int64  `json:"author_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	ProjectID    int64  `json:"project_id"`
	Attachment   string `json:"attachment"`
	LineCode     string `json:"line_code"`
	CommitID     string `json:"commit_id"`
	NoteableID   int64  `json:"noteable_id"`
	System       bool   `json:"system"`
	STDiff       string `json:"st_diff"`
	URL          string `json:"url"`
}
