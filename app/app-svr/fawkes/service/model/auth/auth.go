package auth

import "time"

// Group auth_group 表
type Group struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Operator string    `json:"operator"`
	Ctime    time.Time `json:"ctime"`
	Mtime    time.Time `json:"mtime"`
}

// Item auth_item 表
type Item struct {
	Id          int64     `json:"id"`            //id
	Name        string    `json:"name"`          //权限项名字
	AuthGroupId int64     `json:"auth_group_id"` //关联的权限组
	FeKey       string    `json:"fe_key"`        //前端权限项key
	BeUrl       string    `json:"be_url"`        //后端接口url
	UrlParam    string    `json:"url_param"`     //后端接口参数限制
	Operator    string    `json:"operator"`      //操作人
	IsActive    bool      `json:"is_active"`     //生效状态
	Ctime       time.Time `json:"ctime"`         //创建时间
	Mtime       time.Time `json:"mtime"`         //修改时间
}

// ItemRoleRelation auth_item_role_relation  功能模块与角色关系表
type ItemRoleRelation struct {
	Id            int64     `json:"id"`              //id
	AuthItemId    int64     `json:"auth_item_id"`    //auth_item_id 功能模块外键
	AuthRoleValue int8      `json:"auth_role_value"` //auth_role_value
	Operator      string    `json:"operator"`        //操作人
	Ctime         time.Time `json:"ctime"`           //创建时间
	Mtime         time.Time `json:"mtime"`           //修改时间
}
