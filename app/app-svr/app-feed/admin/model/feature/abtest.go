package feature

import (
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

type ABTestList struct {
	Page *common.Page `json:"page"`
	List []*ABTest    `json:"list"`
}

type ABTest struct {
	ID          int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`             //[ 0] id                                             int                  null: false  primary: true   auto: true   col: int             len: -1      default: []
	TreeID      int        `gorm:"column:tree_id;type:INT;default:0;" json:"tree_id"`                   //[ 1] tree_id                                        int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	KeyName     string     `gorm:"column:key_name;type:VARCHAR;size:50;" json:"key_name"`               //[ 2] key_name                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	ABType      string     `gorm:"column:ab_type;type:VARCHAR;size:32;" json:"ab_type"`                 //[ 3] exp_type                                    varchar(1000)        null: false  primary: false  auto: false  col: varchar         len: 1000    default: []
	Bucket      int        `gorm:"column:bucket;type:INT;default:0;" json:"bucket"`                     //[ 4] bucket                                   int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	Salt        string     `gorm:"column:salt;type:VARCHAR;size:1000;" json:"salt"`                     //[ 5] salt                                    varchar(1000)        null: false  primary: false  auto: false  col: varchar         len: 1000    default: []
	Config      string     `gorm:"column:config;type:VARCHAR;size:1000;" json:"config"`                 //[ 7] config                                     varchar(1000)        null: false  primary: false  auto: false  col: varchar         len: 1000    default: []
	Creator     string     `gorm:"column:creator;type:VARCHAR;size:32;" json:"creator"`                 //[ 8] creator                                        varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	CreatorUID  int        `gorm:"column:creator_uid;type:INT;default:0;" json:"creator_uid"`           //[ 9] creator_uid                                    int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	Modifier    string     `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`               //[10] modifier                                       varchar(32)          null: false  primary: false  auto: false  col: varchar         len: 32      default: []
	ModifierUID int        `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`         //[11] modifier_uid                                   int                  null: false  primary: false  auto: false  col: int             len: -1      default: [0]
	State       string     `gorm:"column:state;type:VARCHAR;size:10;default:off;" json:"state"`         //[12] state                                          varchar(10)          null: false  primary: false  auto: false  col: varchar         len: 10      default: [off]
	Relations   string     `gorm:"column:relations;type:VARCHAR;size:1000;" json:"relations"`           //[16] relations
	Description string     `gorm:"column:description;type:VARCHAR;size:200;" json:"description"`        //[13] description
	Ctime       xtime.Time `gorm:"column:ctime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"ctime"` //[14] ctime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
	Mtime       xtime.Time `gorm:"column:mtime;type:TIMESTAMP;default:CURRENT_TIMESTAMP;" json:"mtime"` //[15] mtime                                          timestamp            null: false  primary: false  auto: false  col: timestamp       len: -1      default: [CURRENT_TIMESTAMP]
}

type ABTestExpConfig struct {
	Group     string `json:"group,omitempty"`
	Start     int    `json:"start,omitempty"`
	End       int    `json:"end,omitempty"`
	Whitelist string `json:"whitelist,omitempty"`
}
