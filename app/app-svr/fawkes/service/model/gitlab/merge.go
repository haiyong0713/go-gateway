package gitlab

import (
	"time"
)

// GitMerge 合并请求记录表
type GitMerge struct {
	Id                int       `json:"id"`                  //id
	MergeId           int       `json:"merge_id"`            //merge的主键 iid
	AppKey            string    `json:"app_key"`             //APP在平台内的唯一标识
	PathWithNamespace string    `json:"path_with_namespace"` //项目地址空间
	State             string    `json:"state"`               //opened,merged,closed
	GitAction         string    `json:"git_action"`          //open,reopen,merge,close
	RequestUser       string    `json:"request_user"`        //mr发起人
	MrTitle           string    `json:"mr_title"`            //合并申请的标题
	MrStartTime       time.Time `json:"mr_start_time"`       //开始merge流程的时间
	MergedTime        time.Time `json:"merged_time"`         //merge成功时间
	Ctime             time.Time `json:"ctime"`               //创建时间
	Mtime             time.Time `json:"mtime"`               //上次修改时间
}
