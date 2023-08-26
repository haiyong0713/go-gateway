package feature

import (
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/tree"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	// treeNode type: app
	NodeApp = 4
)

type AppListRly struct {
	Total int    `json:"total"`
	List  []*App `json:"list"`
}

type App struct {
	TreeID     int    `json:"tree_id"`
	Department string `json:"department"`
	Project    string `json:"project"`
	App        string `json:"app"`
	Count      int    `json:"count"`
	Dimension  string `json:"dimension"`
}

func (a *App) FormTreeNode(node *tree.Node) {
	if node == nil || node.Type != NodeApp {
		return
	}
	a.TreeID = node.TreeID
	a.Department, a.Project, a.App = splitNodePath(node.Path)
	a.Dimension = "service"
}

func splitNodePath(path string) (department, project, app string) {
	parts := strings.Split(path, ".")
	//nolint:gomnd
	if len(parts) < 4 {
		log.Warn("tree node(%s) parts is less than 4.", path)
		return
	}
	var apps []string
	apps = append(apps, parts[1])
	apps = append(apps, parts[2])
	apps = append(apps, parts[3])
	department = parts[1]
	project = parts[2]
	app = strings.Join(apps, ".")
	return
}

type AppPlatReq struct {
	TreeID int `json:"tree_id" form:"tree_id" validate:"required,min=1"`
}

type AppPlatRly struct {
	List map[string][]*AppPlatItem `json:"list"`
}

type AppPlatItem struct {
	IsChosen bool   `json:"is_chosen"`
	PlatName string `json:"plat_name"`
	MobiApp  string `json:"mobi_app"`
}

func (i *AppPlatItem) FormPlat(p *Plat) {
	i.PlatName = p.Name
	i.MobiApp = p.MobiApp
}

type SaveAppReq struct {
	TreeID   int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	MobiApps string `json:"mobi_apps" form:"mobi_apps"`
}

type BuildListReq struct {
	TreeID  int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName string `json:"key_name" form:"key_name"`
	Creator string `json:"creator" form:"creator"`
	Pn      int    `json:"pn" form:"pn" default:"1"`
	Ps      int    `json:"ps" form:"ps" default:"20"`
}

type BuildListRly struct {
	Page *common.Page  `json:"page"`
	List []*BuildLimit `json:"list"`
}

type SaveBuildReq struct {
	ID          int    `json:"id" form:"id" validate:"min=0"`
	TreeID      int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName     string `json:"key_name" form:"key_name" validate:"required"`
	Config      string `json:"config" form:"config"`
	Description string `json:"description" form:"description"`
	//Relations   string `json:"relations" form:"relations"`
}

type BuildConfItem struct {
	MobiApp    string            `json:"mobi_app"`
	Conditions []*BuildCondition `json:"conditions"`
}

type BuildCondition struct {
	Op    string `json:"op"`
	Build int    `json:"Build"`
}

type HandleBuildReq struct {
	ID    int    `json:"id" form:"id" validate:"required,min=1"`
	State string `json:"state" form:"state" validate:"required"`
}

type SwitchTvListReq struct {
	Pn int `json:"pn" form:"pn" default:"1"`
	Ps int `json:"ps" form:"ps" default:"20"`
}

type SwitchTvListReply struct {
	Page *common.Page    `json:"page"`
	List []*SwitchTvItem `json:"list"`
}

type SwitchTvItem struct {
	ID          int         `gorm:"column:id" json:"id"`
	Brand       string      `gorm:"column:brand" json:"brand"`
	Chid        string      `gorm:"column:chid" json:"chid"`
	Model       string      `gorm:"column:model" json:"model"`
	SysVersion  *SysVersion `gorm:"column:sys_version" json:"sys_version"`
	Config      []string    `gorm:"column:config" json:"config"`
	Deleted     int         `gorm:"column:deleted" json:"deleted"`
	Ctime       xtime.Time  `gorm:"column:ctime" json:"ctime"`
	Mtime       xtime.Time  `gorm:"column:mtime" json:"mtime"`
	Description string      `gorm:"description" json:"description"`
}

type SysVersion struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type SwitchTvSaveReq struct {
	ID          int    `json:"id" form:"id" validate:"min=0"`
	Brand       string `json:"brand" form:"brand"`
	Chid        string `json:"chid" form:"chid"`
	Model       string `json:"model" form:"model"`
	SysVersion  string `json:"sys_version" form:"sys_version"`
	Config      string `json:"config" form:"config" validate:"required"`
	Description string `json:"description" form:"description"`
}

type SwitchTvDelReq struct {
	ID int `json:"id" form:"id" validate:"min=1"`
}

type BusinessConfigListReq struct {
	TreeID  int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName string `json:"key_name" form:"key_name"`
	Creator string `json:"creator" form:"creator"`
	Pn      int    `json:"pn" form:"pn" default:"1"`
	Ps      int    `json:"ps" form:"ps" default:"20"`
}

type BusinessConfigReply struct {
	Page *common.Page      `json:"page"`
	List []*BusinessConfig `json:"list"`
}

type BusinessConfigSaveReq struct {
	ID            int    `json:"id" form:"id" validate:"min=0"`
	TreeID        int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName       string `json:"key_name" form:"key_name" validate:"required"`
	Config        string `json:"config" form:"config" validate:"required"`
	Description   string `json:"description" form:"description"`
	Relations     string `json:"relations" form:"relations"`
	WhiteListType string `json:"whitelist_type" form:"whitelist_type"`
	WhiteList     string `json:"whitelist" form:"whitelist"`
}

type BusinessConfigActReq struct {
	ID    int    `json:"id" form:"id" validate:"required,min=1"`
	State string `json:"state" form:"state" validate:"required"`
}

/*
分组实验
*/
type ABTestReq struct {
	TreeID  int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName string `json:"key_name" form:"key_name"`
	Creator string `json:"creator" form:"creator"`
	Pn      int    `json:"pn" form:"pn" default:"1"`
	Ps      int    `json:"ps" form:"ps" default:"20"`
}

type ABTestSaveReq struct {
	ID          int    `json:"id" form:"id" validate:"min=0"`
	TreeID      int    `json:"tree_id" form:"tree_id" validate:"required,min=1"`
	KeyName     string `json:"key_name" form:"key_name" validate:"required"`
	ABType      string `json:"ab_type" form:"ab_type" validate:"required"`
	Bucket      int    `json:"bucket" form:"bucket" validate:"required"`
	Salt        string `json:"salt" form:"salt" validate:"required"`
	Config      string `json:"config" form:"config" validate:"required"`
	Relations   string `json:"relations" form:"relations"`
	Description string `json:"description" form:"description"`
}

type ABTestHandleReq struct {
	ID    int    `json:"id" form:"id" validate:"required,min=1"`
	State string `json:"state" form:"state" validate:"required"`
}
