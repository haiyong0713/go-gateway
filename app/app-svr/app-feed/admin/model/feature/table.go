package feature

import (
	xtime "go-common/library/time"
)

type BuildLimit struct {
	ID          int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`             //[ 0] id                                             int                  null: false  primary: true   auto: true   col: int             len: -1      default: []
	TreeID      int        `gorm:"column:tree_id;type:INT;default:0;" json:"tree_id"`                   //[ 1] tree_id                                        int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	KeyName     string     `gorm:"column:key_name;type:VARCHAR;size:32;" json:"key_name"`               //[ 2] key_name                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	Config      string     `gorm:"column:config;type:VARCHAR;size:1000;" json:"config"`                 //[ 3] config                                         varchar(1000)        null: false  primary: false  auto: false  col: varchar         len: 1000    default: []
	Creator     string     `gorm:"column:creator;type:VARCHAR;size:32;" json:"creator"`                 //[ 4] creator                                        varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	CreatorUID  int        `gorm:"column:creator_uid;type:INT;default:0;" json:"creator_uid"`           //[ 5] creator_uid                                    int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	Modifier    string     `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`               //[ 6] modifier                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	ModifierUID int        `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`         //[ 7] modifier_uid                                   int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	State       string     `gorm:"column:state;type:VARCHAR;size:10;default:off;" json:"state"`         //[ 8] state                                          varchar(10)          null: false  primary: false  auto: false  col: varchar         len: 10      default: [off]
	Ctime       xtime.Time `gorm:"column:ctime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"ctime"` //[ 9] ctime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	Mtime       xtime.Time `gorm:"column:mtime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"mtime"` //[10] mtime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	Description string     `gorm:"column:description;type:VARCHAR;size:200;" json:"description"`        //[11] description
	//Relations   string     `gorm:"column:relations;type:VARCHAR;size:1000;" json:"relations"`           //[12] relations
}

type ServiceAttribute struct {
	ID          int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`             //[ 0] id                                             int                  null: false  primary: true   auto: true   col: int             len: -1      default: []
	TreeID      int        `gorm:"column:tree_id;type:INT;default:0;" json:"tree_id"`                   //[ 1] tree_id                                        int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	MobiApps    string     `gorm:"column:mobi_apps;type:VARCHAR;size:255;" json:"mobi_apps"`            //[ 2] mobi_apps                                      varchar(255)         null: false  primary: false  auto: false  col: varchar         len: 255     default: []
	Modifier    string     `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`               //[ 3] modifier                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	ModifierUID int        `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`         //[ 4] modifier_uid                                   int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	Ctime       xtime.Time `gorm:"column:ctime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"ctime"` //[ 5] ctime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	Mtime       xtime.Time `gorm:"column:mtime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"mtime"` //[ 6] mtime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
}

type SwitchTV struct {
	ID          int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`
	Brand       string     `gorm:"column:brand" json:"brand"`
	Chid        string     `gorm:"column:chid" json:"chid"`
	Model       string     `gorm:"column:model" json:"model"`
	SysVersion  string     `gorm:"column:sys_version" json:"sys_version"`
	Config      string     `gorm:"column:config" json:"config"`
	Deleted     int        `gorm:"column:deleted" json:"deleted"`
	Ctime       xtime.Time `gorm:"column:ctime" json:"ctime"`
	Mtime       xtime.Time `gorm:"column:mtime" json:"mtime"`
	Description string     `gorm:"description" json:"description"`
}

type BusinessConfig struct {
	ID            int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`             //[ 0] id                                             int                  null: false  primary: true   auto: true   col: int             len: -1      default: []
	TreeID        int        `gorm:"column:tree_id;type:INT;default:0;" json:"tree_id"`                   //[ 1] tree_id                                        int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	KeyName       string     `gorm:"column:key_name;type:VARCHAR;size:32;" json:"key_name"`               //[ 2] key_name                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	Config        string     `gorm:"column:config;type:VARCHAR;size:1000;" json:"config"`                 //[ 3] config                                         varchar(1000)        null: false  primary: false  auto: false  col: varchar         len: 1000    default: []
	Description   string     `gorm:"column:description;type:VARCHAR;size:500;" json:"description"`        //[ 4] description
	Relations     string     `gorm:"column:relations;type:VARCHAR;size:1000;" json:"relations"`           //[ 5] relations
	Creator       string     `gorm:"column:creator;type:VARCHAR;size:32;" json:"creator"`                 //[ 6] creator                                        varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	CreatorUID    int        `gorm:"column:creator_uid;type:INT;default:0;" json:"creator_uid"`           //[ 7] creator_uid                                    int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	Modifier      string     `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`               //[ 8] modifier                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	ModifierUID   int        `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`         //[ 9] modifier_uid                                   int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	State         string     `gorm:"column:state;type:VARCHAR;size:10;default:off;" json:"state"`         //[10] state                                          varchar(10)          null: false  primary: false  auto: false  col: varchar         len: 10      default: [off]
	Ctime         xtime.Time `gorm:"column:ctime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"ctime"` //[11] ctime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	Mtime         xtime.Time `gorm:"column:mtime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"mtime"` //[12] mtime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	WhiteListType string     `gorm:"column:whitelist_type;type:VARCHAR;size:32;" json:"whitelist_type"`   //[13] whitelist_type                                 varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	WhiteList     string     `gorm:"column:whitelist;type:VARCHAR;size:500;" json:"whitelist"`            //[13] whitelist                                      varchar(500)          null: false  primary: false  auto: false  col: varchar         len: 500      default: []
}
