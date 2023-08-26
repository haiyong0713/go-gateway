package open

import "time"

const (
	Comma     = ","
	AnyAppKey = "*"
)

// Project 开放接口项目表
type Project struct {
	Id          int64     `json:"id"`           //id
	ProjectName string    `json:"project_name"` //项目名称
	Owner       string    `json:"owner"`        //owner
	Token       string    `json:"token"`        //access token
	Description string    `json:"description"`  //描述
	Applicant   string    `json:"applicant"`    //申请人
	IsActive    bool      `json:"is_active"`    //是否生效
	Ctime       time.Time `json:"ctime"`        //创建时间
	Mtime       time.Time `json:"mtime"`        //修改时间
}

// PathAccess 开放接口权限表
type PathAccess struct {
	Id          int64     `json:"id"`              //id
	ProjectId   int64     `json:"project_id"`      //项目id
	Router      string    `json:"router"`          //接口路由
	AppKey      string    `json:"allowed_app_key"` //可以访问的app_key 逗号隔开
	Description string    `json:"description"`     //描述
	Ctime       time.Time `json:"ctime"`           //创建时间
	Mtime       time.Time `json:"mtime"`           //修改时间
}

// UserProjectRelation 开放接口用户项目关系表
type UserProjectRelation struct {
	Id        int64     `json:"id"`         //id
	UserName  string    `json:"user_name"`  //用户名
	ProjectId int64     `json:"project_id"` //项目id
	Ctime     time.Time `json:"ctime"`      //创建时间
	Mtime     time.Time `json:"mtime"`      //修改时间
}

type TokenClaims struct {
	ProjectName string    `json:"project_name"`
	TimeStamp   time.Time `json:"time_stamp"`
}

// ReqLog 开放接口调用记录表
type ReqLog struct {
	Id          int64     `json:"id"`           //id
	ProjectId   int64     `json:"project_id"`   //项目id
	ProjectName string    `json:"project_name"` //项目名称
	PathId      int64     `json:"path_id"`      //path id
	PathName    string    `json:"path_name"`    //path name
	Ctime       time.Time `json:"ctime"`        //创建时间
	Mtime       time.Time `json:"mtime"`        //修改时间
}
