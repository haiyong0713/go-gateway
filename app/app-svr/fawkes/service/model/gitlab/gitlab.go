package gitlab

import (
	"github.com/xanzy/go-gitlab"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
)

// Tag struct
type Tag struct {
	Name    string  `json:"name"`
	Message string  `json:"message"`
	Target  string  `json:"target"`
	Commit  *Commit `json:"commit,omitempty"`
}

// Branch struct
type Branch struct {
	Name   string  `json:"name"`
	Commit *Commit `json:"commit,omitempty"`
}

// Commit struct
type Commit struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	Title          string    `json:"title"`
	CreatedAt      string    `json:"created_at"`
	ParentIDs      []*string `json:"parent_ids"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	AuthoredDate   string    `json:"authored_date"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedDate  string    `json:"committed_date"`
}

// BranchInfo struct
type BranchInfo struct {
	Name       string         `json:"name"`
	Commit     string         `json:"commit"`
	CommitInfo *gitlab.Commit `json:"commit_info,omitempty"`
}

// BuildPackJobInfo struct
type BuildPackJobInfo struct {
	GitlabProjectID  string
	GitlabPipelineID int
	GitlabJobID      int
}

type GitJobStatus string

const (
	Canceled GitJobStatus = "canceled"
	Failed   GitJobStatus = "failed"
	Pending  GitJobStatus = "pending"
	Running  GitJobStatus = "running"
	Success  GitJobStatus = "success"
	Unknown  GitJobStatus = "unknown"
)

func (p GitJobStatus) Val() int {
	switch p {
	case Canceled:
		return cimdl.CICancel
	case Failed:
		return cimdl.CIFailed
	case Pending:
		return cimdl.CIInWaiting
	case Running:
		return cimdl.CIBuilding
	case Success:
		return cimdl.CISuccess
	default:
		return 0
	}
}

func Convert2PipelineStatus(status int) GitJobStatus {
	switch status {
	case cimdl.CICancel:
		return Canceled
	case cimdl.CIFailed:
		return Failed
	case cimdl.CIInWaiting:
		return Pending
	case cimdl.CIBuilding:
		return Running
	case cimdl.CISuccess:
		return Success
	default:
		return Unknown
	}
}

type BusinessType int

const (
	CI     = BusinessType(0)
	Biz    = BusinessType(1)
	Hotfix = BusinessType(2)
	Tribe  = BusinessType(3)
)

type GitJobStatusChangeInfo struct {
	OriginStatus  GitJobStatus
	CurrentStatus GitJobStatus
	BusinessType  BusinessType
	GitJobId      int64
	Id            int64
}

// MRInfo MR信息  返回参数
type MRInfo struct {
	Repo  string `json:"repo"`
	MrURL string `json:"mr_url"`
}

// CreateRelatedMRResp 创建关联MR  返回参数
type CreateRelatedMRResp struct {
	Message  string    `json:"message"`
	Mainrepo *MRInfo   `json:"mainrepo,omitempty"`
	Subrepos []*MRInfo `json:"subrepos,omitempty"`
}

// BranchCommitResp 远程分支commit 返回参数
type BranchCommitResp struct {
	Message   string `json:"message"`
	CommitSHA string `json:"commit_sha,omitempty"`
}
