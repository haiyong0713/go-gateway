package tribe

import (
	"fmt"
	"time"
)

const (
	Comma     = ","
	TribePath = "tribe"

	PushMavenSuccess = 1
	PushMavenFailed  = -1

	TestEnv = "test"
	ProdEnv = "prod"

	PipelineSuccess = "Success"

	CiInit         = iota
	CiCancel       = -2
	CiFailed       = -1
	CiInWaiting    = 1
	CiBuilding     = 2
	CiBuildSuccess = 3

	CdActive    = 1
	CdNotActive = 0

	PackFilterTypeAll    = 1
	PackFilterTypeTest   = 2
	PackFilterTypeCustom = 0

	PackTask       = "pack"
	PackBundleTask = "biz_build"
)

// Tribe 组件表
type Tribe struct {
	Id          int64     `json:"id" form:"id"`                   //自增ID
	AppKey      string    `json:"app_key" form:"app_key"`         //APP在平台内唯一标识
	Name        string    `json:"name" form:"name"`               //组件名
	CName       string    `json:"c_name" form:"c_name"`           //组件中文名
	Owners      string    `json:"owners" form:"owners"`           //组件管理员
	Description string    `json:"description" form:"description"` //组件描述
	Priority    int64     `json:"priority" form:"priority"`       //优先级
	NoHost      bool      `json:"no_host" form:"no_host"`         //是否不需要宿主
	IsBuildIn   bool      `json:"is_build_in" form:"is_build_in"` //是否内置
	Mtime       time.Time `json:"mtime" form:"mtime"`             //修改时间
	Ctime       time.Time `json:"ctime" form:"ctime"`             //创建时间
}

// BuildPack 组件构建表
type BuildPack struct {
	Id             int64     `json:"id" form:"id"`                             //自增ID
	TribeId        int64     `json:"tribe_id" form:"tribe_id"`                 //组件ID
	GlJobId        int64     `json:"gl_job_id" form:"gl_job_id"`               //gitlab job id
	DepGlJobId     int64     `json:"dep_gl_job_id" form:"dep_gl_job_id"`       //宿主包gitlab job id
	DepFeature     string    `json:"dep_feature" form:"dep_feature"`           //依赖的feature
	GlPrjId        string    `json:"gl_prj_id" form:"gl_prj_id"`               //gitlab project id
	AppId          string    `json:"app_id" form:"app_id"`                     //Android的PackageName或者iOS的BundleID
	GitPath        string    `json:"git_path" form:"git_path"`                 //git 仓库地址
	AppKey         string    `json:"app_key" form:"app_key"`                   //APP在平台内唯一标识
	GitType        int8      `json:"git_type" form:"git_type"`                 //类型: 0 branch,1 tag,2 commit
	GitName        string    `json:"git_name" form:"git_name"`                 //branch:branch名, tag:tag名, commit:short commit ID
	Commit         string    `json:"commit" form:"commit"`                     //构建的commitID
	PkgType        int8      `json:"pkg_type" form:"pkg_type"`                 //类型: 1 debug,2 release,3 enter,4 publish
	Operator       string    `json:"operator" form:"operator"`                 //操作人
	Size           int64     `json:"size" form:"size"`                         //文件大小
	Md5            string    `json:"md5" form:"md5"`                           //包的md5
	PkgPath        string    `json:"pkg_path" form:"pkg_path"`                 //包文件本地地址
	PkgUrl         string    `json:"pkg_url" form:"pkg_url"`                   //包文件url
	MappingUrl     string    `json:"mapping_url" form:"mapping_url"`           //mapping文件url
	BbrUrl         string    `json:"bbr_url" form:"bbr_url"`                   //bbr文件url
	VersionCode    string    `json:"version_code" form:"version_code"`         //版本号
	VersionName    string    `json:"version_name" form:"version_name"`         //版本名字
	State          int8      `json:"state" form:"state"`                       //状态:0正常,1删除
	Status         int8      `json:"status" form:"status"`                     //状态:-2 取消,-1失败,1等待,2打包中,3成功
	DidPush        int8      `json:"did_push" form:"did_push"`                 //是否已经push到cd
	ChangeLog      string    `json:"change_log" form:"change_log"`             //更改记录
	NotifyGroup    int8      `json:"notify_group" form:"notify_group"`         //是否抄送邮件组
	CiEnvVars      string    `json:"ci_env_vars" form:"ci_env_vars"`           //PIPELINE 环境变量
	BuildStartTime time.Time `json:"build_start_time" form:"build_start_time"` //构建开始时间
	BuildEndTime   time.Time `json:"build_end_time" form:"build_end_time"`     //构建结束时间
	Description    string    `json:"description" form:"description"`           //备注
	ErrMsg         string    `json:"err_msg" form:"err_msg"`                   //error msg
	Mtime          time.Time `json:"mtime" form:"mtime"`                       //修改时间
	Ctime          time.Time `json:"ctime" form:"ctime"`                       //创建时间
}

// PackVersion tribe_pack_version tribe包版本表
type PackVersion struct {
	Id          int64     `json:"id"`           //自增ID
	TribeId     int64     `json:"tribe_id"`     //tribe 的主键
	AppId       string    `json:"app_id"`       //app id
	Env         string    `json:"env"`          //环境:test,prod
	VersionCode int64     `json:"version_code"` //宿主版本
	VersionName string    `json:"version_name"` //版本名字
	IsActive    int8      `json:"is_active"`    //是否生效:1生效,-1不生效
	Ctime       time.Time `json:"ctime"`        //创建时间
	Mtime       time.Time `json:"mtime"`        //修改时间
	Operator    string    `json:"operator"`     //操作人
}

// Pack bundle包CD表
type Pack struct {
	Id          int64     `json:"id"`            //自增ID
	AppId       string    `json:"app_id"`        //Android的PackageName或者iOS的BundleID
	AppKey      string    `json:"app_key"`       //APP在平台内唯一标识
	Env         string    `json:"env"`           //环境:test,prod
	TribeId     int64     `json:"tribe_id"`      //组件ID
	GlJobId     int64     `json:"gl_job_id"`     //gitlab job id
	DepGlJobId  int64     `json:"dep_gl_job_id"` //宿主包gitlab job id
	DepFeature  string    `json:"dep_feature"`   //依赖的feature
	VersionId   int64     `json:"version_id"`    //版本id
	GitType     int8      `json:"git_type"`      //类型: 0 branch,1 tag,3 commit
	GitName     string    `json:"git_name"`      //branch:branch名,tag:tag名,commit:short commit ID
	Commit      string    `json:"commit"`        //构建的commitID
	PackType    int8      `json:"pack_type"`     //类型: 0 debug,1 release,2 enter,3 publish
	ChangeLog   string    `json:"change_log"`    //更改记录
	Operator    string    `json:"operator"`      //操作人
	Size        int64     `json:"size"`          //文件大小
	Md5         string    `json:"md5"`           //文件md5
	PackPath    string    `json:"pack_path"`     //包文件本地地址
	PackUrl     string    `json:"pack_url"`      //包文件url
	MappingUrl  string    `json:"mapping_url"`   //mapping文件url
	BbrUrl      string    `json:"bbr_url"`       //bbr文件url
	CdnUrl      string    `json:"cdn_url"`       //cdn url
	Description string    `json:"description"`   //备注
	Sender      string    `json:"sender"`        //推送人
	Mtime       time.Time `json:"mtime"`         //修改时间
	Ctime       time.Time `json:"ctime"`         //创建时间
}

// ConfigFlow 组件流量配置表
type ConfigFlow struct {
	Id             int64     `json:"id"`               //自增ID
	TribeId        int64     `json:"tribe_id"`         //组件id
	Env            string    `json:"env"`              //环境:test,prod
	GlJobId        int64     `json:"gl_job_id"`        //git job id
	Flow           string    `json:"flow"`             //流量配置
	Ctime          time.Time `json:"ctime"`            //创建时间
	Mtime          time.Time `json:"mtime"`            //修改时间
	TribeVersionId int64     `json:"tribe_version_id"` //tribe版本id
	Operator       string    `json:"operator"`         //操作人
}

// ConfigFilter 组件过滤配置表
type ConfigFilter struct {
	Id             int64     `json:"id"`               //自增ID
	TribeId        int64     `json:"tribe_id"`         //tribe 组件id
	TribeVersionId int64     `json:"tribe_version_id"` //版本id
	Env            string    `json:"env"`              //环境:test,prod
	TribePackId    int64     `json:"tribe_pack_id"`    //组件CD表主键ID
	Network        string    `json:"network"`          //网络
	Isp            string    `json:"isp"`              //运营商
	Channel        string    `json:"channel"`          //渠道
	City           string    `json:"city"`             //城市zoneIDs
	Percent        int8      `json:"percent"`          //升级比例
	Salt           string    `json:"salt"`             //盐值,根据version和version_code组合计算
	Device         string    `json:"device"`           //设备ID
	Type           int8      `json:"type"`             //类型:1全量,2内测
	ExcludesSystem string    `json:"excludes_system"`  //排除的系统版本
	Ctime          time.Time `json:"ctime"`            //创建时间
	Mtime          time.Time `json:"mtime"`            //修改时间
	Operator       string    `json:"operator"`         //操作人
}

// PackUpgrade tribe包应用配置表
type PackUpgrade struct {
	Id                int64     `json:"id"`                  //自增ID
	TribeId           int64     `json:"tribe_id"`            //APP在平台内唯一标识
	Env               string    `json:"env"`                 //环境:test,prod
	TribePackId       int64     `json:"tribe_pack_id"`       //tribe_pack表ID
	StartVersionCode  string    `json:"start_version_code"`  //此版本向后的所有版本(包括此版本)
	ChosenVersionCode string    `json:"chosen_version_code"` //指定版本
	Ctime             time.Time `json:"ctime,omitempty"`     //创建时间
	Mtime             time.Time `json:"mtime,omitempty"`     //修改时间
}

// HostRelations 宿主兼容关系表
type HostRelations struct {
	Id             int64     `json:"id"`               //自增ID
	CurrentBuildId int64     `json:"current_build_id"` //当前构建号
	ParentBuildId  int64     `json:"parent_build_id"`  //依赖包的构建号
	AppKey         string    `json:"app_key"`
	Feature        string    `json:"feature"`
	Ctime          time.Time `json:"ctime,omitempty"` //创建时间
	Mtime          time.Time `json:"mtime,omitempty"` //修改时间
}

type MavenData struct {
	AppKey      string
	BundleName  string
	GitlabJobId int64
}

type FlowInfo struct {
	Flows        []string
	GitlabJobIds []int64
}

type TribeApk struct {
	ID             int64         `json:"id"`
	AppKey         string        `json:"app_key"`
	Name           string        `json:"name"`
	Nohost         bool          `json:"no_host"`
	TribeID        int64         `json:"tribe_id"`
	TribeHostJobID int64         `json:"tribe_host_job_id"`
	MD5            string        `json:"md5"`
	ApkCdnURL      string        `json:"apk_cdn_url"`
	Env            string        `json:"env"`
	Priority       int           `json:"priority"`
	BundleVer      int64         `json:"bundle_ver"`
	DepFeature     string        `json:"dep_feature"`
	FilterConfig   *ConfigFilter `json:"filter_config"`
	UpgradeConfig  *PackUpgrade  `json:"upgrade_config"`
}

func (t *TribeApk) String() string {
	return fmt.Sprintf("{name=%v,tribeID=%v,ID=%v,hostID=%v,cdn=%v,env=%v}\n", t.Name, t.TribeID, t.ID, t.TribeHostJobID, t.ApkCdnURL, t.Env)
}

type TribeHostRelation struct {
	CurrentBuildID int64  `json:"current_build_id"`
	ParentBuildID  int64  `json:"parent_build_id"`
	Feature        string `json:"feature"`
	AppKey         string `json:"app_key"`
}

type JobInfo struct {
	ID              int64
	GitlabProjectID string
	GitlabJobID     int64
	Status          int
}
